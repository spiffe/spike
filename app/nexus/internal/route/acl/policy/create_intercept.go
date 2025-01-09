//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package policy

import (
	"fmt"
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	"github.com/spiffe/spike-sdk-go/api/errors"
	"github.com/spiffe/spike-sdk-go/spiffe"
	"github.com/spiffe/spike-sdk-go/validation"

	state "github.com/spiffe/spike/app/nexus/internal/state/base"
	"github.com/spiffe/spike/internal/net"
)

func guardPutPolicyRequest(
	request reqres.PolicyCreateRequest, w http.ResponseWriter, r *http.Request,
) error {
	fmt.Println("in guard put policy request")

	name := request.Name
	spiffeIdPattern := request.SpiffeIdPattern
	pathPattern := request.PathPattern
	permissions := request.Permissions

	spiffeid, err := spiffe.IdFromRequest(r)
	if err != nil {
		fmt.Println("exit 001")

		responseBody := net.MarshalBody(reqres.PolicyCreateResponse{
			Err: data.ErrUnauthorized,
		}, w)
		net.Respond(http.StatusUnauthorized, responseBody, w)
		return err
	}
	err = validation.ValidateSpiffeId(spiffeid.String())
	if err != nil {
		fmt.Println("exit 002")

		responseBody := net.MarshalBody(reqres.PolicyCreateResponse{
			Err: data.ErrUnauthorized,
		}, w)
		net.Respond(http.StatusUnauthorized, responseBody, w)
	}

	allowed := state.CheckAccess(
		spiffeid.String(), "*",
		[]data.PolicyPermission{data.PermissionSuper},
	)
	if !allowed {
		responseBody := net.MarshalBody(reqres.PolicyCreateResponse{
			Err: data.ErrUnauthorized,
		}, w)
		net.Respond(http.StatusUnauthorized, responseBody, w)
		return errors.ErrUnauthorized
	}

	err = validation.ValidateName(name)
	if err != nil {
		fmt.Println("exit 003")

		responseBody := net.MarshalBody(reqres.PolicyCreateResponse{
			Err: data.ErrBadInput,
		}, w)
		net.Respond(http.StatusBadRequest, responseBody, w)
		return err
	}

	err = validation.ValidateSpiffeIdPattern(spiffeIdPattern)
	if err != nil {
		fmt.Println("exit 004", spiffeIdPattern)

		responseBody := net.MarshalBody(reqres.PolicyCreateResponse{
			Err: data.ErrBadInput,
		}, w)
		net.Respond(http.StatusBadRequest, responseBody, w)
		return err
	}

	err = validation.ValidatePathPattern(pathPattern)
	if err != nil {
		fmt.Println("exit 005")

		responseBody :=
			net.MarshalBody(reqres.PolicyCreateResponse{
				Err: data.ErrBadInput,
			}, w)
		net.Respond(http.StatusBadRequest, responseBody, w)
		return err
	}

	err = validation.ValidatePermissions(permissions)
	if err != nil {
		fmt.Println("exit 006")

		responseBody := net.MarshalBody(reqres.PolicyCreateResponse{
			Err: data.ErrBadInput,
		}, w)
		net.Respond(http.StatusBadRequest, responseBody, w)
		return err
	}

	fmt.Println("exit 007")
	return nil
}
