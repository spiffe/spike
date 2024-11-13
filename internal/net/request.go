//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package net

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/spiffe/spike/internal/log"
)

func requestBody(r *http.Request) (bod []byte, err error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	defer func(b io.ReadCloser) {
		if b == nil {
			return
		}
		err = errors.Join(err, b.Close())
	}(r.Body)

	return body, err
}

// ReadRequestBody reads the entire request body from an HTTP request.
// It returns the body as a byte slice if successful. If there is an error reading
// the body or if the body is nil, it writes a 400 Bad Request status to the
// response writer and returns an empty byte slice. Any errors encountered are
// logged.
func ReadRequestBody(r *http.Request, w http.ResponseWriter) []byte {
	body, err := requestBody(r)

	if err != nil {
		log.Log().Info("readRequestBody",
			"msg", "Problem reading request body",
			"err", err.Error())

		w.WriteHeader(http.StatusBadRequest)
		_, err := io.WriteString(w, "")
		if err != nil {
			log.Log().Info("readRequestBody",
				"msg", "Problem writing response",
				"err", err.Error())
		}

		return []byte{}
	}

	if body == nil {
		log.Log().Info("readRequestBody",
			"msg", "No request body.")

		w.WriteHeader(http.StatusBadRequest)
		_, err := io.WriteString(w, "")
		if err != nil {
			log.Log().Info("readRequestBody",
				"msg", "Problem writing response",
				"err", err.Error())
		}
		return []byte{}
	}

	return body
}

// HandleRequestError handles HTTP request errors by writing a 400 Bad Request
// status to the response writer. If err is nil, it returns nil. Otherwise, it
// writes the error status and returns a joined error containing both the original
// error and any error encountered while writing the response.
func HandleRequestError(w http.ResponseWriter, err error) error {
	if err == nil {
		return nil
	}

	w.WriteHeader(http.StatusBadRequest)
	_, writeErr := io.WriteString(w, "")

	return errors.Join(err, writeErr)
}

func HandleRequest[Req any, Res any](
	requestBody []byte,
	w http.ResponseWriter,
	errorResponse Res,
) *Req {
	var request Req
	if err := HandleRequestError(
		w, json.Unmarshal(requestBody, &request),
	); err != nil {
		log.Log().Error("HandleRequest",
			"msg", "Problem unmarshalling request",
			"err", err.Error())

		responseBody := MarshalBody(errorResponse, w)
		if responseBody == nil {
			return nil
		}

		Respond(http.StatusBadRequest, responseBody, w)
		return nil
	}
	return &request
}
