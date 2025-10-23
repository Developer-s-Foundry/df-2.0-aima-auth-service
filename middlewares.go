package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
)

func VerifyGatewayRequest(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		gatewaySecret := os.Getenv("GATEWAY_SECRET_KEY")
		if gatewaySecret == "" {
			writeJSONError(w, http.StatusInternalServerError, "gateway secret not configured")
			return
		}

		timestamp := r.Header.Get("X-Gateway-Timestamp")
		signature := r.Header.Get("X-Gateway-Signature")
		serviceName := r.Header.Get("X-Service-Name") // must be included in signing request

		if timestamp == "" || signature == "" {
			writeJSONError(w, http.StatusUnauthorized, "missing required gateway headers")
			return
		}

		reqTime, err := strconv.ParseInt(timestamp, 10, 64)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid timestamp format")
			return
		}

		if time.Since(time.Unix(reqTime, 0)) > 6*time.Hour {
			writeJSONError(w, http.StatusUnauthorized, "request expired")
			return
		}

		encryptKey := fmt.Sprintf("%s:%s", serviceName, timestamp)
		h := hmac.New(sha256.New, []byte(gatewaySecret))
		h.Write([]byte(encryptKey))
		expectedSig := hex.EncodeToString(h.Sum(nil))

		if !hmac.Equal([]byte(signature), []byte(expectedSig)) {
			writeJSONError(w, http.StatusUnauthorized, "invalid signature")
			return
		}

		next(w, r, ps)
	}
}
