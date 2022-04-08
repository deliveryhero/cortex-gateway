package gateway

import (
	"time"

	"github.com/MicahParks/keyfunc"
	"github.com/cortexproject/cortex/pkg/util/log"
	klog "github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

func newJWKS(cfg Config) *keyfunc.JWKS {
	if cfg.JwksURL == "" {
		return keyfunc.NewGiven(map[string]keyfunc.GivenKey{})
	}
	logger := klog.With(log.Logger)
	options := keyfunc.Options{}
	if cfg.JwksRefreshEnabled {
		level.Debug(logger).Log("msg", "JWKS background refresh enabled", "URL", cfg.JwksURL, "interval", cfg.JwksRefreshInterval, "timeout", cfg.JwksRefreshTimeout)
		options.RefreshInterval = time.Minute * time.Duration(cfg.JwksRefreshInterval)
		options.RefreshTimeout = time.Second * time.Duration(cfg.JwksRefreshTimeout)
		options.RefreshErrorHandler = func(err error) {
			level.Error(logger).Log("msg", "Refreshing JWKS failed", "URL", cfg.JwksURL, "err", err.Error())
		}
	}
	jwks, err := keyfunc.Get(cfg.JwksURL, options)
	if err != nil {
		level.Error(logger).Log("msg", "Create JWKS from url failed", "URL", cfg.JwksURL, "err", err.Error())
		return keyfunc.NewGiven(map[string]keyfunc.GivenKey{})
	}
	return jwks
}
