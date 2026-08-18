package proxy

import (
	"fmt"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
)

func New(rawURL string) (*httputil.ReverseProxy, error) {
	target, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("parse upstream URL %q: %w", rawURL, err)
	}
	if target.Scheme == "" || target.Host == "" {
		return nil, fmt.Errorf("invalid upstream URL %q", rawURL)
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.Rewrite = func(req *httputil.ProxyRequest) {
		req.SetURL(target)
		req.SetXForwarded()
	}

	return proxy, nil
}

func Register(server *gin.Engine, proxyCfgs []ProxyCfg) error {
	for _, proxyCfg := range proxyCfgs {
		proxy, err := New(proxyCfg.Url)

		if err != nil {
			return fmt.Errorf("create %s proxy: %w", proxyCfg.Name, err)
		}

		proxyHandler := gin.WrapH(proxy)

		for _, urlPath := range proxyCfg.UrlPaths {
			server.Any(urlPath, proxyHandler)
		}
	}

	return nil
}
