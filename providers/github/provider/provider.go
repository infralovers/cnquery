// Copyright (c) Mondoo, Inc.
// SPDX-License-Identifier: BUSL-1.1

package provider

import (
	"errors"
	"os"
	"strings"

	"go.mondoo.com/cnquery/v10/llx"
	"go.mondoo.com/cnquery/v10/providers-sdk/v1/inventory"
	"go.mondoo.com/cnquery/v10/providers-sdk/v1/plugin"
	"go.mondoo.com/cnquery/v10/providers-sdk/v1/upstream"
	"go.mondoo.com/cnquery/v10/providers-sdk/v1/vault"
	"go.mondoo.com/cnquery/v10/providers/github/connection"
	"go.mondoo.com/cnquery/v10/providers/github/resources"
)

const ConnectionType = "github"

type Service struct {
	*plugin.Service
}

func Init() *Service {
	return &Service{
		Service: plugin.NewService(),
	}
}

func (s *Service) ParseCLI(req *plugin.ParseCLIReq) (*plugin.ParseCLIRes, error) {
	flags := req.Flags
	if flags == nil {
		flags = map[string]*llx.Primitive{}
	}

	if len(req.Args) == 0 {
		return nil, errors.New("invalid. must specify org, repo, or user")
	}

	conf := &inventory.Config{
		Type:     req.Connector,
		Options:  map[string]string{},
		Discover: &inventory.Discovery{},
	}

	token := ""
	if x, ok := flags["token"]; ok && len(x.Value) != 0 {
		token = string(x.Value)
	}
	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
	}
	if token == "" {
		return nil, errors.New("a valid GitHub token is required, pass --token '<yourtoken>' or set GITHUB_TOKEN environment variable")
	}
	conf.Credentials = append(conf.Credentials, vault.NewPasswordCredential("", token))

	// discovery flags
	discoverTargets := []string{}
	if x, ok := flags["discover"]; ok && len(x.Array) != 0 {
		for i := range x.Array {
			entry := string(x.Array[i].Value)
			discoverTargets = append(discoverTargets, entry)
		}
	} else {
		discoverTargets = []string{"auto"}
	}
	conf.Discover = &inventory.Discovery{Targets: discoverTargets}

	// Do custom flag parsing here
	switch req.Args[0] {
	case "org":
		conf.Options["organization"] = req.Args[1]
	case "user":
		conf.Options["user"] = req.Args[1]
	case "repo":
		conf.Options["repository"] = req.Args[1]
	default:
		return nil, errors.New("invalid GitHub sub-command, supported are: org, user, or repo")
	}

	asset := inventory.Asset{
		Connections: []*inventory.Config{conf},
	}

	return &plugin.ParseCLIRes{Asset: &asset}, nil
}

func (s *Service) MockConnect(req *plugin.ConnectReq, callback plugin.ProviderCallback) (*plugin.ConnectRes, error) {
	return nil, errors.New("mock connect not yet implemented")
}

func (s *Service) Connect(req *plugin.ConnectReq, callback plugin.ProviderCallback) (*plugin.ConnectRes, error) {
	if req == nil || req.Asset == nil {
		return nil, errors.New("no connection data provided")
	}

	conn, err := s.connect(req, callback)
	if err != nil {
		return nil, err
	}

	// We only need to run the detection step when we don't have any asset information yet.
	if req.Asset.Platform == nil {
		if err := s.detect(req.Asset, conn); err != nil {
			return nil, err
		}
	}
	inv, err := s.discover(conn)
	if err != nil {
		return nil, err
	}

	return &plugin.ConnectRes{
		Id:        conn.ID(),
		Name:      conn.Name(),
		Asset:     req.Asset,
		Inventory: inv,
	}, nil
}

func (s *Service) connect(req *plugin.ConnectReq, callback plugin.ProviderCallback) (*connection.GithubConnection, error) {
	if len(req.Asset.Connections) == 0 {
		return nil, errors.New("no connection options for asset")
	}

	asset := req.Asset
	conf := asset.Connections[0]

	runtime, err := s.AddRuntime(func(connId uint32) (*plugin.Runtime, error) {
		conn, err := connection.NewGithubConnection(connId, asset, conf)
		if err != nil {
			return nil, err
		}

		var upstream *upstream.UpstreamClient
		if req.Upstream != nil && !req.Upstream.Incognito {
			upstream, err = req.Upstream.InitClient()
			if err != nil {
				return nil, err
			}
		}

		asset.Connections[0].Id = connId
		return plugin.NewRuntime(
			conn,
			callback,
			req.HasRecording,
			resources.CreateResource,
			resources.NewResource,
			resources.GetData,
			resources.SetData,
			upstream), nil
	})
	if err != nil {
		return nil, err
	}

	return runtime.Connection.(*connection.GithubConnection), err
}

func (s *Service) detect(asset *inventory.Asset, conn *connection.GithubConnection) error {
	asset.Name = conn.Conf.Host

	repoOpt := conn.Conf.Options["repository"]
	ownerOpt := conn.Conf.Options["owner"]
	// try and parse the repo only if the owner isn't explicitly set
	if repoOpt != "" && ownerOpt == "" {
		repoParts := strings.Split(repoOpt, "/")
		if len(repoParts) > 1 {
			conn.Conf.Options["owner"] = repoParts[0]
			conn.Conf.Options["repository"] = repoParts[1]
		}
	}

	platform, err := conn.PlatformInfo()
	if err != nil {
		return err
	}

	asset.Platform = platform
	return nil
}

func (s *Service) discover(conn *connection.GithubConnection) (*inventory.Inventory, error) {
	if conn.Conf.Discover == nil {
		return nil, nil
	}

	runtime, err := s.GetRuntime(conn.ID())
	if err != nil {
		return nil, err
	}

	return resources.Discover(runtime, conn.Conf.Options)
}
