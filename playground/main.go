package main

import (
	"fmt"
	"regexp"
)

func main() {
	spiffeIdPattern := `^spiffe://spike\\.ist/workload/*`
	const validSpiffeIdPattern = `^\^?spiffe://[a-zA-Z0-9.\-*\\]+(/[a-zA-Z0-9._\-*\\]+)*$`

	if match, err := regexp.MatchString(
		validSpiffeIdPattern, spiffeIdPattern); !match {
		fmt.Println("exit 004.1", err)
	} else {
		fmt.Println("exit 004.2")
	}
}
