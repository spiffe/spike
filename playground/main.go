package main

import (
	"encoding/base64"
	"fmt"
	"github.com/cloudflare/circl/group"
	"github.com/cloudflare/circl/secretsharing"
)

//Final key: >>>>>> [96 52 53 102 53 97 97 61 99 106 102 52 101 60 51 99 49 54 52 52 97 51 98 97 53 58 103 96 98 54 111 48]
//Created shares:
//Share 1: ed5c557639aa941ffb765ef9780731bfe8111b9a4a52500836fa32bf42dd5e18
//Share ID:  0x0000000000000000000000000000000000000000000000000000000000000001
//Base64 Encoded Share 1: 7VxVdjmqlB/7dl75eAcxv+gRG5pKUlAINvoyv0LdXhg=
//Decoded Share 1: ed5c557639aa941ffb765ef9780731bfe8111b9a4a52500836fa32bf42dd5e18
//Base64 Encoded Share ID 1: AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAE=
//Decoded Share ID 1: 0000000000000000000000000000000000000000000000000000000000000001
//Share 2: 7a8475873df3c701938257be8ad2301ce20508528c599f2a4500335b272127af
//Share ID:  0x0000000000000000000000000000000000000000000000000000000000000002
//Base64 Encoded Share 2: eoR1hz3zxwGTgle+itIwHOIFCFKMWZ8qRQAzWychJ68=
//Decoded Share 2: 7a8475873df3c701938257be8ad2301ce20508528c599f2a4500335b272127af
//Base64 Encoded Share ID 2: AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAI=
//Decoded Share ID 2: 0000000000000000000000000000000000000000000000000000000000000002
//Share 3: 07ac9598423cf9e32b8e50839d9d2e79dbf8f50ace60ee4c530633f70b64f146
//Share ID:  0x0000000000000000000000000000000000000000000000000000000000000003
//Base64 Encoded Share 3: B6yVmEI8+eMrjlCDnZ0uedv49QrOYO5MUwYz9wtk8UY=
//Decoded Share 3: 07ac9598423cf9e32b8e50839d9d2e79dbf8f50ace60ee4c530633f70b64f146
//Base64 Encoded Share ID 3: AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAM=
//Decoded Share ID 3: 0000000000000000000000000000000000000000000000000000000000000003
//Saved shard for Keeper ID 3 at index 2
//using shares:
//
//Reconstruction successful: true
//Original key:    : [96 52 53 102 53 97 97 61 99 106 102 52 101 60 51 99 49 54 52 52 97 51 98 97 53 58 103 96 98 54 111 48]
//Reconstructed key: [96 52 53 102 53 97 97 61 99 106 102 52 101 60 51 99 49 54 52 52 97 51 98 97 53 58 103 96 98 54 111 48]
//using cloned shares:
//
//Reconstruction successful with cloned shares: true
//Original key:	: [96 52 53 102 53 97 97 61 99 106 102 52 101 60 51 99 49 54 52 52 97 51 98 97 53 58 103 96 98 54 111 48]
//Reconstructed key using cloned shares: [96 52 53 102 53 97 97 61 99 106 102 52 101 60 51 99 49 54 52 52 97 51 98 97 53 58 103 96 98 54 111 48]

func main() {
	g := group.P256

	share1 := "7VxVdjmqlB/7dl75eAcxv+gRG5pKUlAINvoyv0LdXhg="
	decodedShare1Data, err := base64.StdEncoding.DecodeString(share1)
	if err != nil {
		panic("Failed to decode share data: " + err.Error())
	}
	clonedShare1 := secretsharing.Share{
		ID:    g.NewScalar(),
		Value: g.NewScalar(),
	}
	clonedShare1.ID.SetUint64(1)
	err = clonedShare1.Value.UnmarshalBinary(decodedShare1Data)
	if err != nil {
		panic(err)
	}

	share2 := "eoR1hz3zxwGTgle+itIwHOIFCFKMWZ8qRQAzWychJ68="
	decodedShare2Data, err := base64.StdEncoding.DecodeString(share2)
	if err != nil {
		panic("Failed to decode share data: " + err.Error())
	}
	clonedShare2 := secretsharing.Share{
		ID:    g.NewScalar(),
		Value: g.NewScalar(),
	}
	clonedShare2.ID.SetUint64(2)
	err = clonedShare2.Value.UnmarshalBinary(decodedShare2Data)
	if err != nil {
		panic(err)
	}

	var shares []secretsharing.Share
	shares = append(shares, clonedShare1)
	shares = append(shares, clonedShare2)

	// [96 52 53 102 53 97 97 61 99 106 102 52 101 60 51 99 49 54 52 52 97 51 98 97 53 58 103 96 98 54 111 48]

	reconstructed, err := secretsharing.Recover(1, shares)
	if err != nil {
		panic(err)
	}

	binaryRec, _ := reconstructed.MarshalBinary()

	fmt.Println("reconstructed", reconstructed)
	fmt.Println("reconstructed", binaryRec)
}
