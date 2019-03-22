package main

import (
	"github.com/cloudfoundry-community/go-cfclient"
	log "github.com/sirupsen/logrus"
)

// CF --
type CF struct {
	client *cfclient.Client
	entry  *log.Entry
}

// newCF --
func newCF() (*CF, error) {
	res := CF{
		entry: log.WithFields(log.Fields{"module": "cf"}),
	}
	config := &cfclient.Config{
		ApiAddress:   gConfig.CF.URL,
		ClientID:     gConfig.CF.ClientID,
		ClientSecret: gConfig.CF.ClientSecret,
	}
	res.entry.WithFields(log.Fields{
		"url":       gConfig.CF.URL,
		"client_id": gConfig.CF.ClientID,
	}).Debugf("connecting")
	client, err := cfclient.NewClient(config)
	if err != nil {
		res.entry.Errorf("could not connect")
		return nil, err
	}
	res.client = client
	return &res, nil
}

func (s *CF) deleteUser(guid string, name string) {
	entry := s.entry.WithFields(log.Fields{
		"guid": guid,
		"name": name,
	})
	entry.Infof("deleting user")
	if gConfig.DryRun {
		if _, err := s.client.GetUserByGUID(guid); err != nil {
			entry.WithError(err).Warnf("user not find user")
		}
	} else {
		if err := s.client.DeleteUser(guid); err != nil {
			entry.WithError(err).Warnf("could not delete user")
		}
	}
}
