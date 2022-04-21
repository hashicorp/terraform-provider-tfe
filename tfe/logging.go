package tfe

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
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

// IsDebugOrHigher returns whether or not the current log level is debug or trace
func IsDebugOrHigher() bool {
	level := strings.ToUpper(os.Getenv(EnvLog))
	return level == "DEBUG" || level == "TRACE"
}

// RoundTrip is a transport method that logs the request and response if the TF_LOG level is
// TRACE or DEBUG
func (t *loggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if IsDebugOrHigher() {
		reqData, err := httputil.DumpRequestOut(req, true)
		if err == nil {
			log.Printf("[DEBUG] "+logReqMsg, t.name, filterAndPrettyPrintLines(reqData))
		} else {
			log.Printf("[ERROR] %s API Request error: %#v", t.name, err)
		}
	}

	resp, err := t.delegate.RoundTrip(req)
	if err != nil {
		return resp, err
	}

	if IsDebugOrHigher() {
		respData, err := httputil.DumpResponse(resp, true)
		if err == nil {
			log.Printf("[DEBUG] "+logRespMsg, t.name, filterAndPrettyPrintLines(respData))
		} else {
			log.Printf("[ERROR] %s API Response error: %#v", t.name, err)
		}
	}

	return resp, nil
}

// NewLoggingTransport wraps the given transport with a logger that logs request and
// response details
func NewLoggingTransport(name string, t http.RoundTripper) *loggingTransport {
	return &loggingTransport{name, t}
}

// filterAndPrettyPrintLines iterates through a []byte line-by-line,
// redacting any sensitive lines and transforming any lines that are complete json into
// pretty-printed json.
func filterAndPrettyPrintLines(b []byte) string {
	parts := strings.Split(string(b), "\n")
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
