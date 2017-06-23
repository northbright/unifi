package unifi_test

import (
	"context"
	"log"
	"time"

	"github.com/northbright/unifi"
)

func Example() {
	var err error

	unifiURL := "https://192.168.1.10:8443"
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

	// Set timeout to 5 seconds.
	timeout := time.Duration(time.Second * 5)
	// Create context with timeout.
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	exit := make(chan error)
	go func() {
		if err = u.Login(ctx); err != nil {
			exit <- err
			return
		}

		if err = u.Logout(ctx); err != nil {
			exit <- err
			return
		}

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
