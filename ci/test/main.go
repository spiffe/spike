//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package main

import (
	"crypto/rand"
	"fmt"
	"github.com/google/goexpect"
	"log"
	"math/big"
	"regexp"
	"time"
)

func generatePassword(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()_+-=[]"
	password := make([]byte, length)
	for i := range password {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		password[i] = charset[n.Int64()]
	}
	return string(password)
}

func main() {
	password := generatePassword(20)
	timeout := 2 * time.Minute
	spike := "/home/volkan/Desktop/WORKSPACE/spike/spike"

	// Initialize SPIKE.

	child, _, err := expect.Spawn(spike+" init", -1)
	if err != nil {
		log.Fatal(err)
	}
	defer func(child *expect.GExpect) {
		err := child.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(child)

	_, _, err = child.Expect(regexp.MustCompile("Enter admin password:"), timeout)
	if err != nil {
		log.Fatal(err)
	}
	err = child.Send(password + "\n")
	if err != nil {
		log.Fatal(err)
		return
	}

	_, _, err = child.Expect(regexp.MustCompile("Confirm admin password:"), timeout)
	if err != nil {
		log.Fatal(err)
	}
	err = child.Send(password + "\n")
	if err != nil {
		log.Fatal(err)
		return
	}

	_, _, err = child.Expect(regexp.MustCompile("SPIKE system initialization completed."), timeout)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("SPIKE initialized with password: %s\n", password)

	// Log in to SPIKE

	child, _, err = expect.Spawn(spike+" login", -1)
	if err != nil {
		log.Fatal(err)
	}
	defer func(child *expect.GExpect) {
		err := child.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(child)

	_, _, err = child.Expect(regexp.MustCompile("Enter admin password:"), timeout)
	if err != nil {
		log.Fatal(err)
	}
	err = child.Send(password + "\n")
	if err != nil {
		log.Fatal(err)
		return
	}

	_, _, err = child.Expect(regexp.MustCompile("Login successful."), timeout)
	if err != nil {
		log.Fatal(err)
	}

	// Put a secret

	child, _, err = expect.Spawn(spike+" put /tenants/acme/db username=root password=SPIKERocks", -1)
	if err != nil {
		log.Fatal(err)
	}

	_, _, err = child.Expect(regexp.MustCompile("OK"), timeout)
	if err != nil {
		log.Fatal(err)
	}

	// Get the secret

	child, _, err = expect.Spawn(spike+" get /tenants/acme/db", -1)
	if err != nil {
		log.Fatal(err)
	}

	_, _, err = child.Expect(regexp.MustCompile("password: SPIKERocks"), timeout)
}
