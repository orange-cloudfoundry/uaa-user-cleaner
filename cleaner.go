package main

import (
	"encoding/json"
	"github.com/cloudfoundry-community/go-uaa"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

// Cleaner --
type Cleaner struct {
	ldap       *LDAP
	uaa        *UAA
	cf         *CF
	hooks      []Hook
	entry      *log.Entry
	toDelete   []uaa.User
	currentErr error
}

func newCleaner() *Cleaner {
	return &Cleaner{
		entry:      log.WithFields(log.Fields{"module": "cleaner"}),
		toDelete:   []uaa.User{},
		currentErr: nil,
	}
}

func (s *Cleaner) initialize() error {
	ldapHandler, err := newLDAP()
	if err != nil {
		return err
	}
	uaaHandler, err := newUAA()
	if err != nil {
		return err
	}
	cfHandler, err := newCF()
	if err != nil {
		return err
	}
	hooksHandler, err := newHooks()
	if err != nil {
		return err
	}

	s.ldap = ldapHandler
	s.uaa = uaaHandler
	s.cf = cfHandler
	s.hooks = hooksHandler
	return nil
}

func (s *Cleaner) finalize() {
	if s.ldap != nil && s.ldap.conn != nil {
		s.ldap.conn.Close()
	}
}

// Run --
func (s *Cleaner) update() {
	s.entry.Infof("processing users")

	toDelete := []uaa.User{}
	users, err := s.uaa.getUsers()
	if err != nil {
		gMetricNbErrors.Inc()
		s.currentErr = err
		s.entry.Errorf("unable to fetch users from uaa: %s", err)
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
				s.entry.Errorf("unable to query ldap for user '%s': %s", cUser.Username, err)
				gMetricNbErrors.Inc()
				continue
			}
		}
		s.entry.WithFields(log.Fields{
			"user":   cUser.Username,
			"active": active,
		}).Debugf("processed user")
		if !active {
			toDelete = append(toDelete, cUser)
		}
	}

	s.toDelete = toDelete

	gMetricNbUsers.Reset()
	for cOrigin, cValue := range origins {
		gMetricNbUsers.With(prometheus.Labels{
			"origin": cOrigin,
		}).Set(float64(cValue))
	}

	gMetricNbInvalidUsers.Reset()
	for _, cUser := range s.toDelete {
		gMetricNbInvalidUsers.With(prometheus.Labels{
			"origin": cUser.Origin,
			"guid":   cUser.ID,
			"name":   cUser.Username,
		}).Set(1)
	}
}

func (s *Cleaner) prune() {
	for _, cUser := range s.toDelete {
		for _, hook := range s.hooks {
			hook.deleteUser(cUser.ID, cUser.Username)
		}
		s.uaa.deleteUser(cUser.ID, cUser.Username)
		s.cf.deleteUser(cUser.ID, cUser.Username)
	}
}

// Run -
func (s *Cleaner) Run(interval string) {
	duration, _ := time.ParseDuration(interval)
	for {
		timer := prometheus.NewTimer(gMetricDuration)
		gMetricNbErrors.Set(0)
		s.currentErr = nil
		if err := s.initialize(); err != nil {
			s.currentErr = err
			gMetricNbErrors.Set(1)
		} else {
			s.update()
			s.finalize()
			timer.ObserveDuration()
		}
		time.Sleep(duration)
	}
}

// InvalidUser -
type InvalidUser struct {
	Origin   string `json:"origin"`
	Username string `json:"username"`
	GUID     string `json:"guid"`
}

// InvalidUsersResponse -
type InvalidUsersResponse struct {
	StatusOk bool          `json:"status_ok"`
	Error    string        `json:"error"`
	Users    []InvalidUser `json:"users"`
}

// ListInvalidUsers -
func (s *Cleaner) ListInvalidUsers(w http.ResponseWriter, r *http.Request) {
	resp := InvalidUsersResponse{
		StatusOk: true,
		Error:    "",
		Users:    []InvalidUser{},
	}

	if s.currentErr != nil {
		resp.StatusOk = false
		resp.Error = s.currentErr.Error()
	}

	for _, cUser := range s.toDelete {
		resp.Users = append(resp.Users, InvalidUser{
			Origin:   cUser.Origin,
			Username: cUser.Username,
			GUID:     cUser.ID,
		})
	}

	_ = json.NewEncoder(w).Encode(resp)
}

// Local Variables:
// ispell-local-dictionary: "american"
// End:
