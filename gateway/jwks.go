package gateway

import (
	"flag"
	"time"

	"github.com/MicahParks/keyfunc"
	"github.com/cortexproject/cortex/pkg/util/log"
	klog "github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

var (
	jwksURL             string
	jwksRefreshEnabled  bool
	jwksRefreshInterval int
	jwksRefreshTimeout  int
)

func init() {
	flag.StringVar(&jwksURL, "gateway.auth.jwks-url", "", "The URL to load the JWKS (JSON Web Key Set) from")
	flag.BoolVar(&jwksRefreshEnabled, "gateway.auth.jwks-refresh-enabled", false, "Enable the JWKS background refresh. (Default: false)")
	flag.IntVar(&jwksRefreshInterval, "gateway.auth.jwks-refresh-interval", 60, "The JWKS background refresh interval in minutes. (Defaults: 60 minutes)")
	flag.IntVar(&jwksRefreshTimeout, "gateway.auth.jwks-refresh-timeout", 30, "The JWKS background refresh timeout in seconds. (Defaults: 30 seconds)")
}

func newJWKS() *keyfunc.JWKS {
	if jwksURL == "" {
		return keyfunc.NewGiven(map[string]keyfunc.GivenKey{})
	}
	logger := klog.With(log.Logger)
	options := keyfunc.Options{}
	if jwksRefreshEnabled {
		level.Debug(logger).Log("msg", "JWKS background refresh enabled", "URL", jwksURL, "interval", jwksRefreshInterval, "timeout", jwksRefreshTimeout)
		options.RefreshInterval = time.Minute * time.Duration(jwksRefreshInterval)
		options.RefreshTimeout = time.Second * time.Duration(jwksRefreshTimeout)
		options.RefreshErrorHandler = func(err error) {
			level.Error(logger).Log("msg", "Refreshing JWKS failed", "URL", jwksURL, "err", err.Error())
		}
	}
	jwks, err := keyfunc.Get(jwksURL, options)
	if err != nil {
		level.Error(logger).Log("msg", "Create JWKS from url failed", "URL", jwksURL, "err", err.Error())
		return keyfunc.NewGiven(map[string]keyfunc.GivenKey{})
	}
	return jwks
}
