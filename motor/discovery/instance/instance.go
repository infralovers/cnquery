package instance

import (
	"github.com/rs/zerolog/log"
	"go.mondoo.io/mondoo/motor/asset"
	"go.mondoo.io/mondoo/motor/motorid/hostname"
	"go.mondoo.io/mondoo/motor/transports"
	"go.mondoo.io/mondoo/motor/transports/resolver"
)

type Resolver struct{}

func (r *Resolver) Name() string {
	return "Instance Resolver"
}

func (r *Resolver) AvailableDiscoveryTargets() []string {
	return []string{}
}

func (r *Resolver) ParseConnectionURL(url string, opts ...transports.TransportConfigOption) (*transports.TransportConfig, error) {
	return transports.NewTransportFromUrl(url, opts...)
}

func (r *Resolver) Resolve(tc *transports.TransportConfig) ([]*asset.Asset, error) {
	assetInfo := &asset.Asset{
		Connections: []*transports.TransportConfig{tc},
		State:       asset.State_STATE_ONLINE,
	}

	m, err := resolver.New(tc)
	if err != nil {
		return nil, err
	}
	defer m.Close()

	// store detected platform identifier with asset
	assetInfo.PlatformIds = m.Meta.Identifier
	log.Debug().Strs("identifier", assetInfo.PlatformIds).Msg("motor connection")

	// determine platform information
	p, err := m.Platform()
	if err == nil {
		assetInfo.Platform = p
	}

	// use hostname as asset name
	if p != nil && assetInfo.Name == "" {
		// retrieve hostname
		hostname, err := hostname.Hostname(m.Transport, p)
		if err == nil && len(hostname) > 0 {
			assetInfo.Name = hostname
		}
	}

	// use hostname as name if asset name was not explicitly provided
	if assetInfo.Name == "" {
		assetInfo.Name = tc.Host
	}

	return []*asset.Asset{assetInfo}, nil
}
