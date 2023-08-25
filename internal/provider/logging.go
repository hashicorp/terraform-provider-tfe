// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"regexp"
	"strings"
)

type loggingTransport struct {
	name     string
	delegate http.RoundTripper
}

const (
	EnvLog = "TF_LOG"
)

// redactedHeaders is a list of lowercase headers (with trailing colons) that signal that the
// header values should be redacted from logs
var redactedHeaders = []string{"authorization:", "proxy-authorization:"}

// logLevelSet reads the TF_LOG level and ensures it is valid
func logLevelSet() bool {
	level := strings.ToUpper(os.Getenv(EnvLog))
	// Ensure its set to a valid level otherwise will default logging to TRACE
	switch level {
	case "DEBUG", "TRACE", "INFO", "WARN", "ERROR":
		return true
	default:
		return false
	}
}

// RoundTrip is a transport method that logs the request and response if the TF_LOG level is
// TRACE or DEBUG
func (t *loggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	includeBody := !hasSensitiveValues(req)

	// We don't need any logic to handle each specific level as
	// Terraform will log accordingly based on the prefix.
	if logLevelSet() {
		reqData, err := httputil.DumpRequestOut(req, includeBody)
		if err == nil {
			log.Printf("[DEBUG] "+logReqMsg, t.name, filterAndPrettyPrintLines(reqData, includeBody))
		} else {
			log.Printf("[ERROR] %s API Request error: %#v", t.name, err)
		}
	}

	resp, err := t.delegate.RoundTrip(req)
	if err != nil {
		return resp, err
	}

	if logLevelSet() {
		respData, err := httputil.DumpResponse(resp, includeBody)
		if err == nil {
			if strings.Contains(string(respData), "404 Not Found") {
				log.Printf("[WARN] The requested resource at %s %s could not be found. Please ensure no drift occurred by attempting to import the desired resource. It may also be that your token is invalid.", req.Method, req.URL.RequestURI())
			}
			log.Printf("[DEBUG] "+logRespMsg, t.name, filterAndPrettyPrintLines(respData, includeBody))
		} else {
			log.Printf("[ERROR] %s API Response error: %#v", t.name, err)
		}
	}

	return resp, nil
}

func hasSensitiveValues(req *http.Request) bool {
	foundSensitiveVal := false
	if req.Body != nil {
		b, err := io.ReadAll(req.Body)
		if err != nil {
			return true // just to be safe, let's assume there could be a sensitive value
		}

		if regexp.MustCompile(`"sensitive":true`).MatchString(string(b)) {
			foundSensitiveVal = true
		}
		// after done inspecting the body, place back same data we read
		req.Body = io.NopCloser(bytes.NewBuffer(b))
	}
	return foundSensitiveVal
}

// NewLoggingTransport wraps the given transport with a logger that logs request and
// response details
func NewLoggingTransport(name string, t http.RoundTripper) *loggingTransport {
	return &loggingTransport{name, t}
}

// filterAndPrettyPrintLines iterates through a []byte line-by-line,
// redacting any sensitive lines and transforming any lines that are complete json into
// pretty-printed json.
func filterAndPrettyPrintLines(b []byte, includeBody bool) string {
	sanitizedParts := strings.TrimSpace(string(b))
	parts := strings.Split(sanitizedParts, "\n")

	for i, p := range parts {
		for _, check := range redactedHeaders {
			if strings.HasPrefix(strings.ToLower(p), check) {
				// This looks like a sensitive header to redact, so overwrite the entire line
				parts[i] = fmt.Sprintf("%s <REDACTED>", p[0:len(check)])
				continue
			}
		}
		if b := []byte(p); json.Valid(b) {
			var out bytes.Buffer
			_ = json.Indent(&out, b, "", " ") // already checked for validity
			parts[i] = out.String()
		}
	}

	if !includeBody {
		parts = append(parts, "[BODY REDACTED: Due to sensitive values present]")
	}
	return strings.Join(parts, "\n")
}

const logReqMsg = `%s API Request Details:
---[ REQUEST ]---------------------------------------
%s
-----------------------------------------------------`

const logRespMsg = `%s API Response Details:
---[ RESPONSE ]--------------------------------------
%s
-----------------------------------------------------`
