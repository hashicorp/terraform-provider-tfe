// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"sort"
	"strings"
	"sync"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/go-version"
	providerVersion "github.com/hashicorp/terraform-provider-tfe/version"
	svchost "github.com/hashicorp/terraform-svchost"
	"github.com/hashicorp/terraform-svchost/disco"
)

const (
	DefaultHostname = "app.terraform.io"
)

var (
	ErrMissingAuthToken = errors.New("required token could not be found. Please set the token using an input variable in the provider configuration block or by using the TFE_TOKEN environment variable")
	tfeServiceIDs       = []string{"tfe.v2.2"}
)

type ClientConfigMap struct {
	mu     sync.Mutex
	values map[string]*tfe.Client
}

func (c *ClientConfigMap) GetByConfig(config *ClientConfiguration) *tfe.Client {
	if c.mu.TryLock() {
		defer c.Unlock()
	}

	return c.values[config.Key()]
}

func (c *ClientConfigMap) Lock() {
	c.mu.Lock()
}

func (c *ClientConfigMap) Unlock() {
	c.mu.Unlock()
}

func (c *ClientConfigMap) Set(client *tfe.Client, config *ClientConfiguration) {
	if c.mu.TryLock() {
		defer c.Unlock()
	}
	c.values[config.Key()] = client
}

func getTokenFromEnv() string {
	log.Printf("[DEBUG] TFE_TOKEN used for token value")
	return os.Getenv("TFE_TOKEN")
}

func getTokenFromCreds(services *disco.Disco, hostname svchost.Hostname) string {
	log.Printf("[DEBUG] Attempting to fetch token from Terraform CLI configuration for hostname %s...", hostname)
	creds, err := services.CredentialsForHost(hostname)
	if err != nil {
		log.Printf("[DEBUG] Failed to get credentials for %s: %s (ignoring)", hostname, err)
	}
	if creds != nil {
		return creds.Token()
	}
	return ""
}

// GetClient encapsulates the logic for configuring a go-tfe client instance for
// the provider, including fallback to values from environment variables. This
// is useful because we're muxing multiple provider servers together and each
// one needs an identically configured client.
//
// Internally, this function caches configured clients using the specified
// parameters
func GetClient(tfeHost, token string, insecure bool) (*tfe.Client, error) {
	config, err := configure(tfeHost, token, insecure)
	if err != nil {
		return nil, err
	}

	clientCache.Lock()
	defer clientCache.Unlock()

	// Try to retrieve the client from cache
	cached := clientCache.GetByConfig(config)
	if cached != nil {
		return cached, nil
	}

	// Discover the Terraform Enterprise address.
	host, err := config.Services.Discover(config.TFEHost)
	if err != nil {
		return nil, err
	}

	// Get the full Terraform Enterprise service address.
	var address *url.URL
	var discoErr error
	for _, tfeServiceID := range tfeServiceIDs {
		service, err := host.ServiceURL(tfeServiceID)
		if _, ok := err.(*disco.ErrVersionNotSupported); !ok && err != nil {
			return nil, err
		}
		// If discoErr is nil we save the first error. When multiple services
		// are checked and we found one that didn't give an error we need to
		// reset the discoErr. So if err is nil, we assign it as well.
		if discoErr == nil || err == nil {
			discoErr = err
		}
		if service != nil {
			address = service
			break
		}
	}

	if providerVersion.ProviderVersion != "dev" {
		// We purposefully ignore the error and return the previous error, as
		// checking for version constraints is considered optional.
		constraints, _ := host.VersionConstraints(tfeServiceIDs[0], "tfe-provider")

		// First check any constraints we might have received.
		if constraints != nil {
			if err := CheckConstraints(constraints); err != nil {
				return nil, err
			}
		}
	}

	// When we don't have any constraints errors, also check for discovery
	// errors before we continue.
	if discoErr != nil {
		return nil, discoErr
	}

	// Create a new TFE client.
	client, err := tfe.NewClient(&tfe.Config{
		Address:    address.String(),
		Token:      token,
		HTTPClient: config.HTTPClient,
	})
	if err != nil {
		return nil, err
	}

	client.RetryServerErrors(true)
	clientCache.Set(client, config)

	return client, nil
}

// CheckConstraints checks service version constrains against our own
// version and returns rich and informational diagnostics in case any
// incompatibilities are detected.
func CheckConstraints(c *disco.Constraints) error {
	if c == nil || c.Minimum == "" || c.Maximum == "" {
		return nil
	}

	// Generate a parsable constraints string.
	excluding := ""
	if len(c.Excluding) > 0 {
		excluding = fmt.Sprintf(", != %s", strings.Join(c.Excluding, ", != "))
	}
	constStr := fmt.Sprintf(">= %s%s, <= %s", c.Minimum, excluding, c.Maximum)

	// Create the constraints to check against.
	constraints, err := version.NewConstraint(constStr)
	if err != nil {
		return checkConstraintsWarning(err)
	}

	// Create the version to check.
	v, err := version.NewVersion(providerVersion.ProviderVersion)
	if err != nil {
		return checkConstraintsWarning(err)
	}

	// Return if we satisfy all constraints.
	if constraints.Check(v) {
		return nil
	}

	// Find out what action (upgrade/downgrade) we should advice.
	minimum, err := version.NewVersion(c.Minimum)
	if err != nil {
		return checkConstraintsWarning(err)
	}

	maximum, err := version.NewVersion(c.Maximum)
	if err != nil {
		return checkConstraintsWarning(err)
	}

	var excludes []*version.Version
	for _, exclude := range c.Excluding {
		v, err := version.NewVersion(exclude)
		if err != nil {
			return checkConstraintsWarning(err)
		}
		excludes = append(excludes, v)
	}

	// Sort all the excludes.
	sort.Sort(version.Collection(excludes))

	var action, toVersion string
	switch {
	case minimum.GreaterThan(v):
		action = "upgrade"
		toVersion = ">= " + minimum.String()
	case maximum.LessThan(v):
		action = "downgrade"
		toVersion = "<= " + maximum.String()
	case len(excludes) > 0:
		// Get the latest excluded version.
		action = "upgrade"
		toVersion = "> " + excludes[len(excludes)-1].String()
	}

	switch {
	case len(excludes) == 1:
		excluding = fmt.Sprintf(", excluding version %s", excludes[0].String())
	case len(excludes) > 1:
		var vs []string
		for _, v := range excludes {
			vs = append(vs, v.String())
		}
		excluding = fmt.Sprintf(", excluding versions %s", strings.Join(vs, ", "))
	default:
		excluding = ""
	}

	summary := fmt.Sprintf("Incompatible TFE provider version v%s", v.String())
	details := fmt.Sprintf(
		"The configured Terraform Enterprise backend is compatible with TFE provider\n"+
			"versions >= %s, <= %s%s.", c.Minimum, c.Maximum, excluding,
	)

	if action != "" && toVersion != "" {
		summary = fmt.Sprintf("Please %s the TFE provider to %s", action, toVersion)
	}

	// Return the customized and informational error message.
	return fmt.Errorf("%s\n\n%s", summary, details)
}

func checkConstraintsWarning(err error) error {
	return fmt.Errorf(
		"failed to check version constraints: %v\n\n"+
			"checking version constraints is considered optional, but this is an\n"+
			"unexpected error which should be reported",
		err,
	)
}
