package discovery

import (
	"context"
	"fmt"

	"github.com/grandcat/zeroconf"
	"github.com/softwareparatodos/transferlan-plus/core/config"
)

type Announcer struct {
	server *zeroconf.Server
}

func NewAnnouncer() *Announcer {
	return &Announcer{}
}

func (a *Announcer) Start(ctx context.Context, name string, port int, platform string) error {
	if name == "" {
		name = config.ServiceName
	}
	if port == 0 {
		port = config.DefaultPort
	}

	meta := []string{
		"app=TransferLAN+",
		"version=" + config.ProtocolVer,
		"platform=" + platform,
		"role=receiver",
	}

	server, err := zeroconf.Register(name, config.ServiceType, config.ServiceDomain, port, meta, nil)
	if err != nil {
		return fmt.Errorf("no se pudo anunciar el servicio mDNS: %w", err)
	}
	a.server = server

	go func() {
		<-ctx.Done()
		a.Shutdown()
	}()

	return nil
}

func (a *Announcer) Shutdown() {
	if a.server != nil {
		a.server.Shutdown()
		a.server = nil
	}
}
