package gateway

import (
	"flag"
	"fmt"
	"strings"
)

// Config for a gateway
type Config struct {
	DistributorAddress   string
	QueryFrontendAddress string
	RulerAddress         string
	AlertManagerAddress  string
	JwtSecret            string
	ExtraHeaders         string
	TenantName           string
	TenantIDClaim        string
}

// RegisterFlags adds the flags required to config this package's Config struct
func (cfg *Config) RegisterFlags(f *flag.FlagSet) {
	f.StringVar(&cfg.DistributorAddress, "gateway.distributor.address", "", "Upstream HTTP URL for Cortex Distributor")
	f.StringVar(&cfg.QueryFrontendAddress, "gateway.query-frontend.address", "", "Upstream HTTP URL for Cortex Query Frontend")
	f.StringVar(&cfg.RulerAddress, "gateway.ruler.address", "", "Upstream HTTP URL for Cortex Query Frontend")
	f.StringVar(&cfg.AlertManagerAddress, "gateway.alertmanager.address", "", "Upstream HTTP URL for Cortex AlertManager")

	f.StringVar(&cfg.TenantName, "gateway.auth.tenant-name", "", "Tenant name to use when jwt auth disabled")
	f.StringVar(&cfg.JwtSecret, "gateway.auth.jwt-secret", "", "Secret to sign JSON Web Tokens")
	f.StringVar(&cfg.ExtraHeaders, "gateway.auth.jwt-extra-headers", "", "A comma separated list of additional headers to scan for JSON Web Tokens presence")
	f.StringVar(&cfg.TenantIDClaim, "gateway.auth.tenant-id-claim", "tenant_id", "The name of the Tenant ID Claim. Defaults to tenant_id")
}

// Validate given config parameters. Returns nil if everything is fine
func (cfg *Config) Validate() error {
	if cfg.DistributorAddress == "" {
		return fmt.Errorf("you must set -gateway.distributor.address")
	}

	if !strings.HasPrefix(cfg.DistributorAddress, "http") {
		return fmt.Errorf("distributor address must start with a valid scheme (http/https). Given is '%v'", cfg.DistributorAddress)
	}

	if cfg.QueryFrontendAddress == "" {
		return fmt.Errorf("you must set -gateway.query-frontend.address")
	}

	if !strings.HasPrefix(cfg.QueryFrontendAddress, "http") {
		return fmt.Errorf("query frontend address must start with a valid scheme (http/https). Given is '%v'", cfg.QueryFrontendAddress)
	}

	if cfg.RulerAddress == "" {
		return fmt.Errorf("you must set -gateway.ruler.address")
	}

	if !strings.HasPrefix(cfg.RulerAddress, "http") {
		return fmt.Errorf("ruler address must start with a valid scheme (http/https). Given is '%v'", cfg.RulerAddress)
	}
	if cfg.AlertManagerAddress == "" {
		return fmt.Errorf("you must set -gateway.alertmanager.address")
	}

	if !strings.HasPrefix(cfg.AlertManagerAddress, "http") {
		return fmt.Errorf("alertmanager address must start with a valid scheme (http/https). Given is '%v'", cfg.AlertManagerAddress)
	}
	return nil
}
