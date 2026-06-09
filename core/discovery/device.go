package discovery

import (
	"fmt"
	"net"
	"strings"
	"time"
)

type Device struct {
	Name       string
	Host       string
	Port       int
	IPs        []string
	Platform   string
	Version    string
	Discovered time.Time
}

func (d Device) Address() string {
	if len(d.IPs) > 0 {
		return fmt.Sprintf("%s:%d", d.IPs[0], d.Port)
	}
	return fmt.Sprintf("%s:%d", strings.TrimSuffix(d.Host, "."), d.Port)
}

func ipStrings(addrs []net.IP) []string {
	out := make([]string, 0, len(addrs))
	for _, ip := range addrs {
		if ip == nil {
			continue
		}
		out = append(out, ip.String())
	}
	return out
}
