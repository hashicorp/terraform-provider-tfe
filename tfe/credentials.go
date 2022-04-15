package tfe

import (
	"log"
	"os"
	"strings"

	svchost "github.com/hashicorp/terraform-svchost"
	"github.com/hashicorp/terraform-svchost/disco"
)

func collectCredentialsFromEnv() map[svchost.Hostname]string {
	const prefix = "TF_TOKEN_"

	ret := make(map[svchost.Hostname]string)
	for _, ev := range os.Environ() {
		eqIdx := strings.Index(ev, "=")
		if eqIdx < 0 {
			continue
		}
		name := ev[:eqIdx]
		value := ev[eqIdx+1:]
		if !strings.HasPrefix(name, prefix) {
			continue
		}
		rawHost := name[len(prefix):]

		// We accept double underscores in place of hyphens because hyphens are not valid
		// identifiers in most shells and are therefore hard to set.
		// This is unambiguous with replacing single underscores below because
		// hyphens are not allowed at the beginning or end of a label and therefore
		// odd numbers of underscores will not appear together in a valid variable name.
		rawHost = strings.ReplaceAll(rawHost, "__", "-")

		// We accept underscores in place of dots because dots are not valid
		// identifiers in most shells and are therefore hard to set.
		// Underscores are not valid in hostnames, so this is unambiguous for
		// valid hostnames.
		rawHost = strings.ReplaceAll(rawHost, "_", ".")

		// Because environment variables are often set indirectly by OS
		// libraries that might interfere with how they are encoded, we'll
		// be tolerant of them being given either directly as UTF-8 IDNs
		// or in Punycode form, normalizing to Punycode form here because
		// that is what the Terraform credentials helper protocol will
		// use in its requests.
		//
		// Using ForDisplay first here makes this more liberal than Terraform
		// itself would usually be in that it will tolerate pre-punycoded
		// hostnames that Terraform normally rejects in other contexts in order
		// to ensure stored hostnames are human-readable.
		dispHost := svchost.ForDisplay(rawHost)
		hostname, err := svchost.ForComparison(dispHost)
		if err != nil {
			// Ignore invalid hostnames
			continue
		}

		ret[hostname] = value
	}

	return ret
}

// hostTokenFromFallbackSources returns a token credential by searching for a hostname-specific
// environment variable, a TFE_TOKEN, or a CLI config credentials block. The host parameter
// is expected to be in the "comparison" form, for example, hostnames containing non-ASCII
// characters like "café.fr" should be expressed as "xn--caf-dma.fr". If the variable based
// on the hostname is not defined, nil is returned.
//
// Hyphen and period characters are allowed in environment variable names, but are not valid POSIX
// variable names. However, it's still possible to set variable names with these characters using
// utilities like env or docker. Variable names may have periods translated to underscores and
// hyphens translated to double underscores in the variable name.
// For the example "café.fr", you may use the variable names "TF_TOKEN_xn____caf__dma_fr",
// "TF_TOKEN_xn--caf-dma_fr", or "TF_TOKEN_xn--caf-dma.fr"
func hostTokenFromFallbackSources(hostname svchost.Hostname, services *disco.Disco) string {
	token, ok := collectCredentialsFromEnv()[hostname]

	if ok {
		log.Printf("[DEBUG] TF_TOKEN_... used for token value for host %s", hostname)
	} else {
		// If a token wasn't set in the host-specific variable, try and fetch it
		// from the environment or from Terraform's CLI configuration or configured credential helper.
		if os.Getenv("TFE_TOKEN") != "" {
			log.Printf("[DEBUG] TFE_TOKEN used for token value")
			return os.Getenv("TFE_TOKEN")
		} else if services != nil {
			log.Printf("[DEBUG] Attempting to fetch token from Terraform CLI configuration for hostname %s...", hostname)
			creds, err := services.CredentialsForHost(hostname)
			if err != nil {
				log.Printf("[DEBUG] Failed to get credentials for %s: %s (ignoring)", hostname, err)
			}
			if creds != nil {
				token = creds.Token()
			}
		}
	}

	return token
}
