package unifi

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
)

var (
	rawURLs = map[string]string{
		"login":  "/api/login",
		"logout": "/api/logout",
	}
)

// Unifi provides functions to call Unifi APIs.
type Unifi struct {
	userName string
	password string
	baseURL  *url.URL
	urls     map[string]*url.URL
	jar      *cookiejar.Jar
}

// New creates a new Unifi.
//
// Params:
//     unifiURL: Unifi Controller's URL. E.g. https://10.0.1.100:8443
//     userName: User name of Unifi Controller.
//     password: Password of Unifi Controller.
func New(unifiURL, userName, password string) (*Unifi, error) {
	var err error

	u := &Unifi{}
	if u.baseURL, err = url.Parse(unifiURL); err != nil {
		err = fmt.Errorf("Parse Unifi URL error: %v", err)
		return u, err
	}

	u.urls = map[string]*url.URL{}
	for k, v := range rawURLs {
		refURL, _ := url.Parse(v)
		u.urls[k] = u.baseURL.ResolveReference(refURL)
	}

	u.userName = userName
	u.password = password

	if u.jar, err = cookiejar.New(nil); err != nil {
		err = fmt.Errorf("cookiejar.New() error: %v", err)
		return u, err
	}

	return u, err
}

// ParseJSON parses the JSON returned by Unifi APIs.
//
// Params:
//     b: Bytes returned by Unifi APIs which contains JSON string.
// Return:
//     map[string]interface{} as parsed JSON object.
//     true or false if "rc" is "ok".
func ParseJSON(b []byte) (map[string]interface{}, bool, error) {
	var err error
	m := map[string]interface{}{}

	if err = json.Unmarshal(b, &m); err != nil {
		err = fmt.Errorf("json.Unmarshal() error: %v", err)
		return m, false, err
	}

	if _, ok := m["meta"]; !ok {
		err = fmt.Errorf("'meta' does not exist in returned JSON.")
		return m, false, err
	}

	meta, ok := m["meta"].(map[string]interface{})
	if !ok {
		err = fmt.Errorf("meta type: %T, can't convert meta to map[string]interface{}", m["meta"])
		return m, false, err
	}

	rc, ok := meta["rc"].(string)
	if !ok {
		err = fmt.Errorf("rc type: %T, can't convert rc to string", meta["rc"])
		return m, false, err
	}

	return m, rc == "ok", err
}

// Login() logins Unifi Controller.
func (u *Unifi) Login() error {
	var err error

	// POST data is in JSON format.
	login := struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{
		u.userName,
		u.password,
	}

	b, err := json.Marshal(login)
	if err != nil {
		err = fmt.Errorf("json.Marshal() error: %v", err)
		return err
	}

	buf := bytes.NewBuffer(b)

	// Login.
	req, err := http.NewRequest("POST", u.urls["login"].String(), buf)
	if err != nil {
		err = fmt.Errorf("NewRequest error: %v", err)
		return err
	}

	req.Header.Set("Accept", "*/*")
	req.Header.Set("Content-Type", "application/json")

	tr := &http.Transport{
		// Skip cert verify.
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	resp, err := client.Do(req)
	if err != nil {
		err = fmt.Errorf("client.Do() error: %v", err)
		return err
	}
	defer resp.Body.Close()

	b, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		err = fmt.Errorf("ReadAll() error: %v", err)
		return err
	}

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("Login Unifi failed. Response: %v", string(b))
		return err
	}

	respCookies := resp.Cookies()
	// Set cookie for cookiejar manually.
	u.jar.SetCookies(u.baseURL, respCookies)

	return err
}

// Logout logouts Unifi Controller.
func (u *Unifi) Logout() error {
	var err error

	// Logout.
	// Method: POST.
	req, err := http.NewRequest("POST", u.urls["logout"].String(), nil)
	if err != nil {
		err = fmt.Errorf("NewRequest error: %v", err)
		return err
	}

	req.Header.Set("Accept", "*/*")

	tr := &http.Transport{
		// Skip cert verify.
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	resp, err := client.Do(req)
	if err != nil {
		err = fmt.Errorf("client.Do() error: %v", err)
		return err
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		err = fmt.Errorf("ReadAll() error: %v", err)
		return err
	}

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("Logout Unifi failed. Response: %v", b)
		return err
	}

	return err
}
