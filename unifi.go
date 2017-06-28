package unifi

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
)

const (
	defaultSite = "default"
)

var (
	debugMode = false
	rawURLs   = map[string]string{
		"login":  "/api/login",
		"logout": "/api/logout",
		"stamgr": "/api/s/$site/cmd/stamgr",
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

// SetDebugMode sets debug mode for package unifi.
func SetDebugMode(f bool) {
	debugMode = f
}

// IsDebugMode returns if it's in debug mode or not.
func IsDebugMode() bool {
	return debugMode
}

// logFnResult outputs the result of the function.
//
// params:
//     funcName: function name.
//     err: result of function.
func logFnResult(funcName string, err error) {
	if !debugMode {
		return
	}

	if err != nil {
		log.Printf("%v() error: %v", funcName, err)
		return
	}

	log.Printf("%v() ok", funcName)
}

// New creates a new Unifi.
//
// Params:
//     unifiURL: Unifi Controller's URL. E.g. https://10.0.1.100:8443
//     userName: User name of Unifi Controller.
//     password: Password of Unifi Controller.
func New(unifiURL, userName, password string) (*Unifi, error) {
	var err error

	defer logFnResult("New", err)

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

	defer logFnResult("ParseJSON", err)

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

// Login logins Unifi Controller.
//
// Params:
//     ctx: Parent context. You may use context.Background() to create an empty context.
//          See http://godoc.org/context for more info.
func (u *Unifi) Login(ctx context.Context) error {
	var err error

	defer logFnResult("Login", err)

	// POST data is in JSON format.
	args := struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{
		u.userName,
		u.password,
	}

	b, err := json.Marshal(args)
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
	// Get a copy of req with its context changed to ctx.
	req = req.WithContext(ctx)

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

	if debugMode {
		b, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			err = fmt.Errorf("ReadAll() error: %v", err)
			return err
		}
		log.Printf("Login() response: %v", string(b))
	}

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("response status code: %v", resp.StatusCode)
		return err
	}

	respCookies := resp.Cookies()
	// Set cookie for cookiejar manually.
	u.jar.SetCookies(u.baseURL, respCookies)

	return err
}

// Logout logouts Unifi Controller.
//
// Params:
//     ctx: Parent context. You may use context.Background() to create an empty context.
//          See http://godoc.org/context for more info.
func (u *Unifi) Logout(ctx context.Context) error {
	var err error

	defer logFnResult("Logout", err)

	// Logout.
	// Method: POST.
	req, err := http.NewRequest("POST", u.urls["logout"].String(), nil)
	if err != nil {
		err = fmt.Errorf("NewRequest error: %v", err)
		return err
	}
	// Get a copy of req with its context changed to ctx.
	req = req.WithContext(ctx)

	req.Header.Set("Accept", "*/*")

	tr := &http.Transport{
		// Skip cert verify.
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr, Jar: u.jar}

	resp, err := client.Do(req)
	if err != nil {
		err = fmt.Errorf("client.Do() error: %v", err)
		return err
	}
	defer resp.Body.Close()

	if debugMode {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			err = fmt.Errorf("ReadAll() error: %v", err)
			return err
		}
		log.Printf("Logout() response: %v", string(b))
	}

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("response status code: %v", resp.StatusCode)
		return err
	}

	respCookies := resp.Cookies()
	// Set cookie for cookiejar manually.
	u.jar.SetCookies(u.baseURL, respCookies)

	return err
}

// AuthorizeGuestWithQos authorizes guest by MAC, time and set qos.
//
// Params:
//     ctx: Parent context. You may use context.Background() to create an empty context.
//          See http://godoc.org/context for more info.
//     site: Site name. It's **NOT** the "Site Name"(just description) in Unifi GUI.
//           If you only have 1 site. Just use "default" or leave it empty.
//           If you've created new sites, follow this to get the site name:
//           https://github.com/northbright/Notes/blob/master/Software/unifi/use-compass-to-explore-mongodb-of-unifi/use-compass-to-explore-mongodb-of-unifi.md
//     mac: MAC address of guest to be authorized. It's in "aa:bb:cc:dd:ee:ff" format.
//     min: Timeout in minutes.
//     down: Max download speed in KB.
//     up: Max upload speed in KB.
//     quota: Quota in MB.
func (u *Unifi) AuthorizeGuestWithQos(ctx context.Context, site, mac string, min, down, up, quota int) error {
	var err error

	defer logFnResult("AuthorizeGuestWithQos", err)

	if site == "" {
		site = "default"
	}

	args := map[string]string{}
	args["cmd"] = "authorize-guest"
	args["mac"] = mac
	args["minutes"] = strconv.Itoa(min)

	if down > 0 {
		args["down"] = strconv.Itoa(down)
	}

	if up > 0 {
		args["up"] = strconv.Itoa(up)
	}

	if quota > 0 {
		args["bytes"] = strconv.Itoa(quota)
	}

	b, err := json.Marshal(args)
	if err != nil {
		err = fmt.Errorf("json.Marshal() error: %v", err)
		return err
	}

	buf := bytes.NewBuffer(b)

	urlStr := u.urls["stamgr"].String()
	// Replace $site with real site.
	urlStr = strings.Replace(urlStr, "$site", site, -1)

	if debugMode {
		log.Printf("AuthorizeGuestWithQos(): POST URL: %v", urlStr)
		log.Printf("AuthorizeGuestWithQos(): POST data: %v", string(b))
	}

	// Authorize Guest.
	req, err := http.NewRequest("POST", urlStr, buf)
	if err != nil {
		err = fmt.Errorf("NewRequest error: %v", err)
		return err
	}
	// Get a copy of req with its context changed to ctx.
	req = req.WithContext(ctx)

	req.Header.Set("Accept", "*/*")
	req.Header.Set("Content-Type", "application/json")

	tr := &http.Transport{
		// Skip cert verify.
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr, Jar: u.jar}

	resp, err := client.Do(req)
	if err != nil {
		err = fmt.Errorf("client.Do() error: %v", err)
		return err
	}
	defer resp.Body.Close()

	if debugMode {
		b, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			err = fmt.Errorf("ReadAll() error: %v", err)
			return err
		}
		log.Printf("AuthorizeGuest() response: %v", string(b))
	}

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("response status code: %v", resp.StatusCode)
		return err
	}

	respCookies := resp.Cookies()
	// Set cookie for cookiejar manually.
	u.jar.SetCookies(u.baseURL, respCookies)

	return err
}

// AuthorizeGuest authorizes guest by MAC, time.It's a wrapper of AuthorizeGuestWithQos.
//
// Params:
//     ctx: Parent context. You may use context.Background() to create an empty context.
//          See http://godoc.org/context for more info.
//     site: Site name. It's **NOT** the "Site Name"(just description) in Unifi GUI.
//           If you only have 1 site. Just use "default" or leave it empty.
//           If you've created new sites, follow this to get the site name:
//           https://github.com/northbright/Notes/blob/master/Software/unifi/use-compass-to-explore-mongodb-of-unifi/use-compass-to-explore-mongodb-of-unifi.md
//
//     mac: MAC address of guest to be authorized. It's in "aa:bb:cc:dd:ee:ff" format.
//     min: Timeout in minutes.
func (u *Unifi) AuthorizeGuest(ctx context.Context, site, mac string, min int) error {
	return u.AuthorizeGuestWithQos(ctx, site, mac, min, 0, 0, 0)
}
