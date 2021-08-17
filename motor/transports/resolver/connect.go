package resolver

import (
	"github.com/cockroachdb/errors"
	"github.com/rs/zerolog/log"
	"go.mondoo.io/mondoo/motor"
	"go.mondoo.io/mondoo/motor/asset"
	"go.mondoo.io/mondoo/motor/transports"
	"go.mondoo.io/mondoo/motor/vault"
)

func EstablishConnection(tc *transports.TransportConfig, credentialFn func(cred *vault.Credential) (*vault.Credential, error), idDetectors []string, insecure bool, record bool) (*motor.Motor, error) {
	log.Debug().Str("connection", tc.ToUrl()).Bool("insecure", insecure).Msg("establish connection to asset")
	// overwrite connection specific insecure with global insecure
	if insecure {
		tc.Insecure = insecure
	}

	if record {
		tc.Record = true
	}

	return NewMotorConnection(tc, credentialFn, idDetectors...)
}

func OpenAssetConnection(assetInfo *asset.Asset, credentialFn func(cred *vault.Credential) (*vault.Credential, error), record bool) (*motor.Motor, error) {
	if assetInfo == nil {
		return nil, errors.New("asset is not defined")
	}

	// connect to the platform
	if len(assetInfo.Connections) == 0 {
		return nil, errors.New("no connection provided for asset " + assetInfo.Name)
	}

	// TODO: we may want to allow multiple connection trials later
	tc := assetInfo.Connections[0]

	// use connection host as default
	if assetInfo.Name == "" {
		assetInfo.Name = tc.Host
	}

	// some transports have their own kind/runtime information already
	// NOTE: going forward we may want to enforce that assets have at least kind and runtime information
	if assetInfo.Platform != nil {
		tc.Kind = assetInfo.Platform.Kind
		tc.Runtime = assetInfo.Platform.Runtime
	}

	// parse reference id and restore options
	if len(assetInfo.PlatformIds) > 0 {
		tc.PlatformId = assetInfo.PlatformIds[0]
	}

	return EstablishConnection(tc, credentialFn, nil, tc.Insecure, record)
}

func OpenAssetConnections(assetInfo *asset.Asset, credentialFn func(cred *vault.Credential) (*vault.Credential, error), record bool) ([]*motor.Motor, error) {
	if assetInfo == nil {
		return nil, errors.New("asset is not defined")
	}

	// connect to the platform
	if len(assetInfo.Connections) == 0 {
		return nil, errors.New("no connection provided for asset " + assetInfo.Name)
	}

	// TODO: we may want to allow multiple connection trials later
	connections := []*motor.Motor{}
	for ci := range assetInfo.Connections {
		tc := assetInfo.Connections[ci]

		// use connection host as default
		if assetInfo.Name == "" {
			assetInfo.Name = tc.Host
		}

		// some transports have their own kind/runtime information already
		// NOTE: going forward we may want to enforce that assets have at least kind and runtime information
		if assetInfo.Platform != nil {
			tc.Kind = assetInfo.Platform.Kind
			tc.Runtime = assetInfo.Platform.Runtime
		}

		// parse reference id and restore options
		if len(assetInfo.PlatformIds) > 0 {
			tc.PlatformId = assetInfo.PlatformIds[0]
		}

		m, err := EstablishConnection(tc, credentialFn, assetInfo.IdDetector, tc.Insecure, record)
		if err != nil {
			return nil, err
		}

		connections = append(connections, m)
	}
	return connections, nil
}
