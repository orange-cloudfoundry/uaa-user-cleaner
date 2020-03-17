package main

import (
	"github.com/cloudfoundry-community/go-uaa"
	log "github.com/sirupsen/logrus"
)

// UAA --
type UAA struct {
	api   *uaa.API
	entry *log.Entry
}

// newUAA --
func newUAA() (*UAA, error) {
	res := UAA{
		entry: log.WithFields(log.Fields{"module": "uaa"}),
	}
	api, err := res.newAPI()
	if err != nil {
		return nil, err
	}
	res.api = api
	return &res, nil
}
func (s *UAA) newAPI() (*uaa.API, error) {
	s.entry.WithFields(log.Fields{
		"url":                 gConfig.UAA.URL,
		"client_id":           gConfig.UAA.ClientID,
		"skip_ssl_validation": gConfig.UAA.SkipSslValidation,
	}).Debugf("connecting")

	client, err := uaa.New(
		gConfig.UAA.URL,
		uaa.WithClientCredentials(
			gConfig.UAA.ClientID,
			gConfig.UAA.ClientSecret,
			uaa.JSONWebToken,
		),
		uaa.WithSkipSSLValidation(gConfig.UAA.SkipSslValidation),
		uaa.WithVerbosity(false),
	)

	if err != nil {
		s.entry.WithError(err).Errorf("unable to connect")
		return nil, err
	}
	return client, err
}

func (s *UAA) getUsers() ([]uaa.User, error) {
	s.entry.Debugf("fething user list")
	users, err := s.api.ListAllUsers("", "", "", uaa.SortAscending)
	if err != nil {
		s.entry.WithError(err).Errorf("unable to fetch user list")
		return nil, err
	}
	return users, nil
}

func (s *UAA) deleteUser(guid string, name string) {
	entry := s.entry.WithFields(log.Fields{
		"guid":    guid,
		"name":    name,
		"dry_run": gConfig.DryRun,
	})
	entry.Infof("deleting user")
	if gConfig.DryRun {
		_, err := s.api.GetUser(guid)
		if err != nil {
			entry.WithError(err).Warnf("could not find user")
		}
	} else {
		_, err := s.api.DeleteUser(guid)
		if err != nil {
			gMetricNbErrors.Inc()
			entry.WithError(err).Errorf("could not delete user")
		}
	}
}
