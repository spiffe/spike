package main

import (
	"fmt"
	"github.com/spiffe/spike-sdk-go/security/mem"
)

func main() {
	s := []int{1, 2, 3}

	fmt.Println(len(s))

	for _, v := range s {
		println(v)
	}

	mem.ClearRawBytes(&s)

	fmt.Println(len(s))

	for _, v := range s {
		println(v)
	}
}
