package unifi_test

import (
	"log"

	"github.com/northbright/unifi"
)

func Example() {
	var err error

	unifiURL := "https://192.168.1.10:8443"
	userName := "admin"
	password := "admin"

	defer func() {
		if err != nil {
			log.Printf("%v", err)
		}
	}()

	u, err := unifi.New(unifiURL, userName, password)
	if err != nil {
		return
	}

	if err = u.Login(); err != nil {
		return
	}
	defer u.Logout()
	// Output:
}
