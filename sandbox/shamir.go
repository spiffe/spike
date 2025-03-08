package main

import (
	"encoding/base64"
	"fmt"
	"github.com/cloudflare/circl/group"
	shamir "github.com/cloudflare/circl/secretsharing"
	"github.com/spiffe/spike-sdk-go/crypto"
)

func main() {
	rootKeyBase64 := "F6bJW0SWNQSv7nW6k8V7O6RgESsKYFyCuSpfYjTDJGk="
	rootKey, _ := base64.StdEncoding.DecodeString(rootKeyBase64)

	g := group.P256

	rootSecret := g.NewScalar()

	err := rootSecret.UnmarshalBinary(rootKey)
	if err != nil {
		panic(err)
	}

	t := uint(2 - 1)

	reader := crypto.NewDeterministicReader(rootKey)
	ss := shamir.New(reader, t, rootSecret)

	sss := ss.Share(3)

	for _, s := range sss {
		fmt.Println("-----")
		fmt.Println("ID", s.ID)
		fmt.Println(s.ID.IsEqual(g.NewScalar().SetUint64(2)))
		fmt.Println("Value", s.Value)
	}
}
