package uaa

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"errors"

	"golang.org/x/oauth2"
)

func (a *API) doJSON(method string, url *url.URL, body io.Reader, response interface{}, needsAuthentication bool) error {
	return a.doJSONWithHeaders(method, url, nil, body, response, needsAuthentication)
}

func (a *API) doJSONWithHeaders(method string, url *url.URL, headers map[string]string, body io.Reader, response interface{}, needsAuthentication bool) error {
	req, err := http.NewRequest(method, url.String(), body)
	if err != nil {
		return err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	bytes, err := a.doAndRead(req, needsAuthentication)
	if err != nil {
		return err
	}

	if response != nil {
		if err := json.Unmarshal(bytes, response); err != nil {
			return parseError(err, url.String(), bytes)
		}
	}

	return nil
}

func (a *API) doAndRead(req *http.Request, needsAuthentication bool) ([]byte, error) {
	req.Header.Add("Accept", "application/json")
	req.Header.Add("X-Identity-Zone-Id", a.ZoneID)
	switch req.Method {
	case http.MethodPut, http.MethodPost, http.MethodPatch:
		req.Header.Add("Content-Type", "application/json")
	}
	if a.Verbose {
		logRequest(req)
	}
	a.ensureTimeout()
	var (
		resp *http.Response
		err  error
	)
	if needsAuthentication {
		if a.AuthenticatedClient == nil {
			return nil, errors.New("doAndRead: the HTTPClient cannot be nil")
		}
		a.ensureTransport(a.AuthenticatedClient)
		resp, err = a.AuthenticatedClient.Do(req)
	} else {
		a.ensureTransport(a.UnauthenticatedClient)
		resp, err = a.UnauthenticatedClient.Do(req)
	}

	if err != nil {
		if a.Verbose {
			fmt.Printf("%v\n\n", err)
		}

		return nil, requestError(req.URL.String())
	}

	if a.Verbose {
		logResponse(resp)
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		if a.Verbose {
			fmt.Printf("%v\n\n", err)
		}
		return nil, unknownError()
	}

	if !is2XX(resp.StatusCode) {
		return nil, requestError(req.URL.String())
	}
	return bytes, nil
}

func (a *API) ensureTimeout() {
	if a.AuthenticatedClient != nil && a.AuthenticatedClient.Timeout == 0 {
		a.AuthenticatedClient.Timeout = time.Second * 120
	}

	if a.UnauthenticatedClient != nil && a.UnauthenticatedClient.Timeout == 0 {
		a.UnauthenticatedClient.Timeout = time.Second * 120
	}
}

func (a *API) ensureTransport(c *http.Client) {
	if c == nil {
		return
	}
	switch t := c.Transport.(type) {
	case *oauth2.Transport:
		b, ok := t.Base.(*http.Transport)
		if !ok {
			return
		}
		if b.TLSClientConfig == nil && !a.SkipSSLValidation {
			return
		}
		if b.TLSClientConfig == nil {
			b.TLSClientConfig = &tls.Config{}
		}
		b.TLSClientConfig.InsecureSkipVerify = a.SkipSSLValidation
	case *tokenTransport:
		if t.underlyingTransport.TLSClientConfig == nil && !a.SkipSSLValidation {
			return
		}
		if t.underlyingTransport.TLSClientConfig == nil {
			t.underlyingTransport.TLSClientConfig = &tls.Config{}
		}
		t.underlyingTransport.TLSClientConfig.InsecureSkipVerify = a.SkipSSLValidation
	case *http.Transport:
		if t.TLSClientConfig == nil && !a.SkipSSLValidation {
			return
		}
		if t.TLSClientConfig == nil {
			t.TLSClientConfig = &tls.Config{}
		}
		t.TLSClientConfig.InsecureSkipVerify = a.SkipSSLValidation
	}
}
