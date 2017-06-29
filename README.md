# Unifi

[![Build Status](https://travis-ci.org/northbright/unifi.svg?branch=master)](https://travis-ci.org/northbright/unifi)
[![Go Report Card](https://goreportcard.com/badge/github.com/northbright/unifi)](https://goreportcard.com/report/github.com/northbright/unifi)

package unifi is a [Golang](https://golang.org) SDK for [UBNT](https://www.ubnt.com/) [Unifi](https://unifi-sdn.ubnt.com/) APIs to interact with Unifi Controller.

Currently, it focuses on guest authorization which is useful to implment customized portal server.

#### Features
* Login / Logout Unifi Controller.
* Authorize / Unauthorize Guest.

#### Timeout and Cancelation Support
* API's first parameter is a [context.Context](http://godoc.org/context#Context).
*  [context.Context](http://godoc.org/context#Context) make it possible to support timeout, deadline, cancelation while calling APIs.
* If you don't need timeout and cancelation, use [context.Background()](http://godoc.org/context#Background) to create a non-nil, empty Context.

#### Requirement
* [Go 1.7](golang.org/doc/go1.7) or higher is required(Go 1.7 moves the `golang.org/x/net/context` package into the standard library as context).

#### Example

```
func Example() {
        var err error

        unifiURL := "https://192.168.1.56:8443"
        userName := "admin"
        password := "admin"

        defer func() {
                if err != nil {
                        log.Printf("error: %v", err)
                }
        }()

        // New an Unifi instance.
        u, err := unifi.New(unifiURL, userName, password)
        if err != nil {
                return
        }

        // Set debug mode to true to output debug messages(default is false).
        unifi.SetDebugMode(true)
        // Set timeout to 5 seconds.
        timeout := time.Duration(time.Second * 5)
        // Create context with timeout.
        ctx, cancel := context.WithTimeout(context.Background(), timeout)
        defer cancel()

        exit := make(chan error)
        go func() {
                // Login
                if err = u.Login(ctx); err != nil {
                        exit <- err
                        return
                }

                // Logout before return if login successfully.
                defer func() {
                        exit <- u.Logout(ctx)
                }()

                mac := "aa:bb:cc:dd:ee:ff" // MAC address.
                min := 5                   // time to authorize in minutes.

                // Authorize guest with MAC and time.
                if err = u.AuthorizeGuest(ctx, "default", mac, min); err != nil {
                        exit <- err
                        return
                }


                // Unauthorize guest with MAC.
                /*if err = u.UnAuthorizeGuest(ctx, "default", mac); err != nil {
                        exit <- err
                        return
                }*/

                exit <- nil
        }()

        select {
        case err = <-exit:
                log.Printf("goroutine exited")
                return

        case <-ctx.Done():
                err = ctx.Err()
                return
        }
        // Output:
}
```

#### Documentation
* [API References](http://godoc.org/github.com/northbright/unifi)

#### License
* [MIT License](LICENSE)

