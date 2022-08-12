package winrm

import (
	"bytes"
	"errors"
	"os"
	"time"

	"github.com/masterzen/winrm"
	"github.com/rs/zerolog/log"
	"github.com/spf13/afero"
	"go.mondoo.io/mondoo/motor/providers"
	"go.mondoo.io/mondoo/motor/providers/winrm/cat"
	"go.mondoo.io/mondoo/motor/vault"
)

var _ providers.Transport = (*WinrmTransport)(nil)

func VerifyConfig(endpoint *providers.TransportConfig) (*winrm.Endpoint, error) {
	if endpoint.Backend != providers.ProviderType_WINRM {
		return nil, errors.New("only winrm backend for winrm transport supported")
	}

	winrmEndpoint := &winrm.Endpoint{
		Host:     endpoint.Host,
		Port:     int(endpoint.Port),
		Insecure: endpoint.Insecure,
		HTTPS:    true,
		Timeout:  time.Duration(0),
	}

	return winrmEndpoint, nil
}

func DefaultConfig(endpoint *winrm.Endpoint) *winrm.Endpoint {
	// use default port if port is 0
	if endpoint.Port <= 0 {
		endpoint.Port = 5986
	}

	if endpoint.Port == 5985 {
		log.Warn().Msg("winrm port 5985 is using http communication instead of https, passwords are not encrypted")
		endpoint.HTTPS = false
	}

	if os.Getenv("WINRM_DISABLE_HTTPS") == "true" {
		log.Warn().Msg("WINRM_DISABLE_HTTPS is set, winrm is using http communication instead of https, passwords are not encrypted")
		endpoint.HTTPS = false
	}

	return endpoint
}

// New creates a winrm client and establishes a connection to verify the connection
func New(tc *providers.TransportConfig) (*WinrmTransport, error) {
	// ensure all required configs are set
	winrmEndpoint, err := VerifyConfig(tc)
	if err != nil {
		return nil, err
	}

	// set default config if required
	winrmEndpoint = DefaultConfig(winrmEndpoint)

	params := winrm.DefaultParameters
	params.TransportDecorator = func() winrm.Transporter { return &winrm.ClientNTLM{} }

	// search for password secret
	c, err := vault.GetPassword(tc.Credentials)
	if err != nil {
		return nil, errors.New("missing password for winrm transport")
	}

	client, err := winrm.NewClientWithParameters(winrmEndpoint, c.User, string(c.Secret), params)
	if err != nil {
		return nil, err
	}

	// test connection
	log.Debug().Str("user", c.User).Str("host", tc.Host).Msg("winrm> connecting to remote shell via WinRM")
	shell, err := client.CreateShell()
	if err != nil {
		return nil, err
	}

	err = shell.Close()
	if err != nil {
		return nil, err
	}

	log.Debug().Msg("winrm> connection established")
	return &WinrmTransport{
		Endpoint: winrmEndpoint,
		Client:   client,
		kind:     tc.Kind,
		runtime:  tc.Runtime,
	}, nil
}

type WinrmTransport struct {
	Endpoint *winrm.Endpoint
	Client   *winrm.Client
	kind     providers.Kind
	runtime  string
	fs       afero.Fs
}

func (t *WinrmTransport) RunCommand(command string) (*providers.Command, error) {
	log.Debug().Str("command", command).Str("transport", "winrm").Msg("winrm> run command")

	stdoutBuffer := &bytes.Buffer{}
	stderrBuffer := &bytes.Buffer{}

	mcmd := &providers.Command{
		Command: command,
		Stdout:  stdoutBuffer,
		Stderr:  stderrBuffer,
	}

	// Note: winrm does not return err of the command was executed with a non-zero exit code
	exitCode, err := t.Client.Run(command, stdoutBuffer, stderrBuffer)
	if err != nil {
		log.Error().Err(err).Str("command", command).Msg("could not execute winrm command")
		return mcmd, err
	}

	mcmd.ExitStatus = exitCode
	return mcmd, nil
}

func (t *WinrmTransport) FileInfo(path string) (providers.FileInfoDetails, error) {
	fs := t.FS()
	afs := &afero.Afero{Fs: fs}
	stat, err := afs.Stat(path)
	if err != nil {
		return providers.FileInfoDetails{}, err
	}

	uid := int64(-1)
	gid := int64(-1)
	mode := stat.Mode()

	return providers.FileInfoDetails{
		Mode: providers.FileModeDetails{mode},
		Size: stat.Size(),
		Uid:  uid,
		Gid:  gid,
	}, nil
}

func (t *WinrmTransport) FS() afero.Fs {
	if t.fs == nil {
		t.fs = cat.New(t)
	}
	return t.fs
}

func (t *WinrmTransport) Close() {
	// nothing to do yet
}

func (t *WinrmTransport) Capabilities() providers.Capabilities {
	return providers.Capabilities{
		providers.Capability_RunCommand,
		providers.Capability_File,
	}
}

func (t *WinrmTransport) Kind() providers.Kind {
	return t.kind
}

func (t *WinrmTransport) Runtime() string {
	return t.runtime
}

func (t *WinrmTransport) PlatformIdDetectors() []providers.PlatformIdDetector {
	return []providers.PlatformIdDetector{
		providers.HostnameDetector,
		providers.CloudDetector,
	}
}
