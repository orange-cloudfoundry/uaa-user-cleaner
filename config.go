package main

import (
	"errors"
	"github.com/cloudfoundry-community/gautocloud"
	"github.com/cloudfoundry-community/gautocloud/connectors/generic"
	"net/url"
	"strconv"
	"time"
)

// CFConfig -
type CFConfig struct {
	URL               string `cloud:"url"`
	ClientID          string `cloud:"client_id"`
	ClientSecret      string `cloud:"client_secret"`
	SkipSslValidation bool   `cloud:"skip_ssl_validation"`
}

// Validate -
func (t *CFConfig) Validate() error {
	if t.URL == "" {
		return errors.New("missing mandatory configuration key cf.url")
	}
	if t.ClientID == "" {
		return errors.New("missing mandatory configuration key cf.client_id")
	}
	if t.ClientSecret == "" {
		return errors.New("missing mandatory configuration key cf.client_secret")
	}
	return nil
}

// UAAConfig -
type UAAConfig struct {
	URL               string `cloud:"url"`
	ClientID          string `cloud:"client_id"`
	ClientSecret      string `cloud:"client_secret"`
	SkipSslValidation bool   `cloud:"skip_ssl_validation"`
}

// Validate -
func (t *UAAConfig) Validate() error {
	if t.URL == "" {
		return errors.New("missing mandatory configuration key uaa.url")
	}
	if t.ClientID == "" {
		return errors.New("missing mandatory configuration key uaa.client_id")
	}
	if t.ClientSecret == "" {
		return errors.New("missing mandatory configuration key uaa.client_secret")
	}
	return nil
}

// LDAPConfig -
type LDAPConfig struct {
	URL               string `cloud:"url"`
	BindUser          string `cloud:"bind_user"`
	BindPassword      string `cloud:"bind_password"`
	SearchBase        string `cloud:"search_base"`
	ValidFilter       string `cloud:"valid_filter"`
	SkipSslValidation bool   `cloud:"skip_ssl_validation"`
}

// Validate -
func (w *LDAPConfig) Validate() error {
	if w.URL == "" {
		return errors.New("missing mandatory configuration key ldap.url")
	}
	if w.BindUser == "" {
		return errors.New("missing mandatory configuration key ldap.bind_user")
	}
	if w.BindPassword == "" {
		return errors.New("missing mandatory configuration key ldap.bind_password")
	}
	if w.SearchBase == "" {
		return errors.New("missing mandatory configuration key ldap.search_base")
	}
	if w.ValidFilter == "" {
		return errors.New("missing mandatory configuration key ldap.valid_filter")
	}
	if _, err := url.Parse(w.URL); err != nil {
		return errors.New("invalid url in configuration key ldap.url")
	}
	return nil
}

// GetAddress
func (w *LDAPConfig) GetAddress() (string, int, bool) {
	url, _ := url.Parse(w.URL)

	url.Hostname()
	port := url.Port()
	isTLS := (url.Scheme == "ldaps")
	if port == "" {
		port = "389"
		if isTLS {
			port = "636"
		}
	}
	iport, _ := strconv.Atoi(port)
	return url.Hostname(), iport, isTLS
}

// LogConfig -
type LogConfig struct {
	Level   string `cloud:"level"`
	JSON    bool   `cloud:"json"`
	NoColor bool   `cloud:"no_color"`
}

// WebConfig -
type WebConfig struct {
	Listen  string `cloud:"listen"`
	SSLCert string `cloud:"ssl_cert"`
	SSLKey  string `cloud:"ssl_key"`
}

// Validate -
func (w *WebConfig) Validate() error {
	if w.Listen == "" {
		return errors.New("missing mandatory configuration key web.listen")
	}
	return nil
}

// Config -
type Config struct {
	DryRun   bool       `cloud:"dry_run"`
	Interval string     `cloud:"interval"`
	CF       CFConfig   `cloud:"cf"`
	LDAP     LDAPConfig `cloud:"ldap"`
	UAA      UAAConfig  `cloud:"uaa"`
	Log      LogConfig  `cloud:"log"`
	Web      WebConfig  `cloud:"web"`
}

// Validate -
func (c *Config) Validate() error {
	if err := c.LDAP.Validate(); err != nil {
		return err
	}
	if err := c.UAA.Validate(); err != nil {
		return err
	}
	if err := c.CF.Validate(); err != nil {
		return err
	}
	if err := c.Web.Validate(); err != nil {
		return err
	}

	if c.Interval == "" {
		c.Interval = "4h"
	}
	if _, err := time.ParseDuration(c.Interval); err != nil {
		return errors.New("invalid time duration in configuration key interval")
	}
	return nil
}

func init() {
	gautocloud.RegisterConnector(generic.NewConfigGenericConnector(Config{}))
}

// Local Variables:
// ispell-local-dictionary: "american"
// End:
