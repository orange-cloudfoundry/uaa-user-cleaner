package main

import (
	"github.com/cloudfoundry-community/go-uaa"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

// Cleaner --
type Cleaner struct {
	ldap  *LDAP
	uaa   *UAA
	cf    *CF
	entry *log.Entry
}

func newCleaner() (*Cleaner, error) {
	ldap, err := newLDAP()
	if err != nil {
		gMetricNbErrors.Inc()
		return nil, err
	}
	uaa, err := newUAA()
	if err != nil {
		gMetricNbErrors.Inc()
		return nil, err
	}
	cf, err := newCF()
	if err != nil {
		gMetricNbErrors.Inc()
		return nil, err
	}

	return &Cleaner{
		ldap:  ldap,
		uaa:   uaa,
		cf:    cf,
		entry: log.WithFields(log.Fields{"module": "cleaner"}),
	}, nil
}

// Close --
func (s *Cleaner) Close() {
	if s.ldap != nil && s.ldap.conn != nil {
		s.ldap.conn.Close()
	}
}

// Run --
func (s *Cleaner) Run() {
	timer := prometheus.NewTimer(gMetricDuration)
	s.entry.Infof("processing users")

	toDelete := []uaa.User{}
	users, err := s.uaa.getUsers()
	if err != nil {
		gMetricNbErrors.Inc()
		return
	}

	origins := map[string]int{}
	for _, cUser := range users {
		val, ok := origins[cUser.Origin]
		if ok {
			origins[cUser.Origin] = val + 1
		} else {
			origins[cUser.Origin] = 1
		}
		active := true
		if cUser.Origin == "ldap" {
			active, err = s.ldap.isActive(cUser.Username)
			if err != nil {
				gMetricNbErrors.Inc()
				continue
			}
		}
		s.entry.WithFields(log.Fields{
			"user":   cUser.Username,
			"active": active,
		}).Infof("processed user")
		if !active {
			toDelete = append(toDelete, cUser)
		}
	}

	for cOrigin, cValue := range origins {
		gMetricNbUsers.With(prometheus.Labels{
			"origin": cOrigin,
		}).Set(float64(cValue))
	}

	for _, cUser := range toDelete {
		s.uaa.deleteUser(cUser.ID, cUser.Username)
		s.cf.deleteUser(cUser.ID, cUser.Username)
	}
	timer.ObserveDuration()
}

// Local Variables:
// ispell-local-dictionary: "american"
// End:
