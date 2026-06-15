package ai

import (
	"net/http"
	"net/url"
	"strings"
	"time"
)

func newProviderHTTPClient(config ServiceConfig) *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if proxyAddr := strings.TrimSpace(config.HTTPProxyAddress); proxyAddr != "" {
		if proxyURL, err := url.Parse(proxyAddr); err == nil {
			transport.Proxy = http.ProxyURL(proxyURL)
		}
	}
	return &http.Client{
		Transport: transport,
		Timeout:   normalizedProviderTimeout(config.Timeout),
	}
}

func normalizedProviderTimeout(timeout time.Duration) time.Duration {
	if timeout <= 0 {
		return 30 * time.Second
	}
	return timeout
}
