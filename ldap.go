package main

import (
	"crypto/tls"
	"fmt"
	"strings"
	"time"

	ldap "github.com/go-ldap/ldap/v3"
	log "github.com/sirupsen/logrus"
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

func (s *LDAP) parseTime(value string) (time.Time, error) {
	formatStr := "20060102150405Z"
	return time.Parse(formatStr, value)
}

func (s *LDAP) isActive(username string) (bool, error) {
	maxLastModified := time.Now().Add(24 * time.Hour * time.Duration(gConfig.LDAP.LastModifiedMaxDays))
	maxLastModifiedStr := maxLastModified.Format("20060102150405Z")
	search := gConfig.LDAP.ValidFilter
	search = strings.ReplaceAll(search, "\n", "")
	search = strings.ReplaceAll(search, "{username}", username)
	search = strings.ReplaceAll(search, "{maxModifiedTimestamp}", maxLastModifiedStr)

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

	if len(res.Entries) == 0 {
		s.entry.WithField("username", username).Debugf("user not found")
		return false, nil
	}

	if len(res.Entries) != 1 {
		err := fmt.Errorf("non-unique result for username %s", username)
		s.entry.WithError(err).Errorf("user not unique")
		return true, err
	}

	entry := res.Entries[0]
	s.entry.
		WithField("dn", entry.DN).
		Debugf("user found")
	return true, nil
}
