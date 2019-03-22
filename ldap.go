package main

import (
	"crypto/tls"
	"fmt"
	log "github.com/sirupsen/logrus"
	"gopkg.in/ldap.v3"
	"strings"
)

// LDAP --
type LDAP struct {
	conn  *ldap.Conn
	entry *log.Entry
}

// NewCleaner --
func newLDAP() (*LDAP, error) {
	res := LDAP{
		entry: log.WithFields(log.Fields{"module": "ldap"}),
	}
	conn, err := res.newClient()
	if err != nil {
		return nil, err
	}
	res.conn = conn
	return &res, nil
}
func (s *LDAP) newClient() (*ldap.Conn, error) {
	address, port, isTLS := gConfig.LDAP.GetAddress()
	s.entry.WithFields(log.Fields{
		"address": address,
		"port":    port,
	}).Debugf("connecting")
	conn, err := ldap.Dial("tcp", fmt.Sprintf("%s:%d", address, port))
	if err != nil {
		s.entry.WithError(err).Errorf("could not connect")
		return nil, err
	}

	if isTLS {
		s.entry.WithFields(log.Fields{
			"skip_ssl_validation": gConfig.LDAP.SkipSslValidation,
		}).Debugf("negotiating TLS")
		err = conn.StartTLS(&tls.Config{
			InsecureSkipVerify: gConfig.LDAP.SkipSslValidation,
			ServerName:         address,
		})
		if err != nil {
			s.entry.WithError(err).Errorf("could not start TLS")
			return nil, err
		}
	}

	s.entry.WithFields(log.Fields{
		"bind_user": gConfig.LDAP.BindUser,
	}).Debugf("authenticating")
	err = conn.Bind(gConfig.LDAP.BindUser, gConfig.LDAP.BindPassword)
	if err != nil {
		s.entry.WithError(err).Errorf("unable to authenticate")
		return nil, err
	}

	return conn, nil
}

func (s *LDAP) isActive(username string) (bool, error) {
	search := strings.ReplaceAll(gConfig.LDAP.ValidFilter, "{0}", username)
	req := ldap.NewSearchRequest(
		gConfig.LDAP.SearchBase,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases,
		0, 0, false,
		search,
		[]string{"dn", "cn"},
		nil,
	)

	res, err := s.conn.Search(req)
	if err != nil {
		s.entry.WithError(err).Errorf("error on ldap search")
		return false, err
	}
	for _, entry := range res.Entries {
		s.entry.WithField("dn", entry.DN).Debugf("found user")
	}
	if len(res.Entries) != 0 {
		return true, nil
	}
	return false, nil
}
