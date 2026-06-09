package discovery

import (
	"context"
	"strings"
	"time"

	"github.com/grandcat/zeroconf"
	"github.com/softwareparatodos/transferlan-plus/core/config"
)

type Browser struct{}

func NewBrowser() *Browser {
	return &Browser{}
}

func (b *Browser) Browse(ctx context.Context, timeout time.Duration) ([]Device, error) {
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		return nil, err
	}

	entries := make(chan *zeroconf.ServiceEntry)
	devices := make([]Device, 0)
	seen := map[string]bool{}

	browseCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	err = resolver.Browse(browseCtx, config.ServiceType, config.ServiceDomain, entries)
	if err != nil {
		return nil, err
	}

	for {
		select {
		case entry := <-entries:
			if entry == nil {
				continue
			}
			key := entry.Instance + entry.HostName
			if seen[key] {
				continue
			}
			seen[key] = true

			device := Device{
				Name:       entry.Instance,
				Host:       entry.HostName,
				Port:       entry.Port,
				IPs:        ipStrings(entry.AddrIPv4),
				Platform:   txtValue(entry.Text, "platform"),
				Version:    txtValue(entry.Text, "version"),
				Discovered: time.Now(),
			}

			if len(device.IPs) == 0 {
				device.IPs = ipStrings(entry.AddrIPv6)
			}
			devices = append(devices, device)

		case <-browseCtx.Done():
			return devices, nil
		}
	}
}

func txtValue(txt []string, key string) string {
	prefix := key + "="
	for _, item := range txt {
		if strings.HasPrefix(item, prefix) {
			return strings.TrimPrefix(item, prefix)
		}
	}
	return ""
}
