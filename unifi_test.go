package unifi_test

import (
	"context"
	"log"
	"time"

	"github.com/northbright/unifi"
)

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

	u, err := unifi.New(unifiURL, userName, password)
	if err != nil {
		return
	}

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

		down := 2048 // download speed in KB.
		up := 512    // upload speed in KB.
		quota := 20  // Quota limit in MB.

		// Authorize guest with MAC, time, download speed, upload speed and quota.
		if err = u.AuthorizeGuestWithQos(ctx, "default", mac, min, down, up, quota); err != nil {
			exit <- err
			return
		}

		// Unauthorize guest with MAC.
		/*if err = u.UnAuthorizeGuest(ctx, "default", mac); err != nil {
			exit <- err
			return
		}*/

		// List sta
		s, err := u.ListSta(ctx, "default")
		if err != nil {
			exit <- err
			return
		}
		log.Printf("STA: %v", s)

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
