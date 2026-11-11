package routes

import (
	"fmt"
	"html/template"
	"log"
	"net"
	"os"
	"strings"
)

func ParseTrustedProxies(raw string) ([]string, error) {
	var proxies []string
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if _, _, err := net.ParseCIDR(part); err != nil {
			if net.ParseIP(part) == nil {
				return nil, fmt.Errorf("invalid trusted proxy %q", part)
			}
		}
		proxies = append(proxies, part)
	}

	return proxies, nil
}

func (r Router) trustProxies(proxies []string) error {
	if len(proxies) == 0 {
		proxies = nil
	}

	return r.engine.SetTrustedProxies(proxies)
}

func (r Router) Config() {
	proxies, err := ParseTrustedProxies(os.Getenv("TRUSTED_PROXIES"))
	if err != nil {
		log.Fatal(err)
	}
	if err := r.trustProxies(proxies); err != nil {
		log.Fatal(err)
	}

	r.engine.SetFuncMap(template.FuncMap{"asset": assetURL(newAssetIndex(staticRoot))})
	r.engine.LoadHTMLGlob("web/template/*")
}
