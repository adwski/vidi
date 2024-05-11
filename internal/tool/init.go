package tool

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"os"

	userapi "github.com/adwski/vidi/internal/api/user/client"
	videoapi "github.com/adwski/vidi/internal/api/video/grpc/userside/pb"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// initialize initializes application up to the point
// that is possible with existing state (config).
func (t *Tool) initialize() {
	if t.state == nil {
		t.state = newState(t.dir)
	}
	if err := t.state.load(); err != nil {
		t.logger.Warn("cannot load state", zap.Error(err))
		// avoid side effects
		t.state = newState(t.dir)
	}
	if !t.state.noEndpoint() {
		t.err = t.initClients(t.state.Endpoint)
	}
	t.cycleViews()
}

// initStateDir creates state dir if it not exists.
func initStateDir() (string, error) {
	dir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot identify home dir: %w", err)
	}
	if err = os.MkdirAll(dir+stateDir, stateDirPerm); err != nil {
		return "", fmt.Errorf("cannot create state directory: %w", err)
	}
	return dir + stateDir, nil
}

// initClients initializes video api and user api clients using provided ViDi endpoint.
func (t *Tool) initClients(ep string) error {
	rCfg, err := t.getRemoteConfig(ep)
	if err != nil {
		return err
	}

	// GRPC client is always spawned with tls creds
	// We're using CA from remote config
	cp := x509.NewCertPool()
	if !cp.AppendCertsFromPEM(rCfg.vidiCA) {
		return fmt.Errorf("credentials: failed to append certificates")
	}
	creds := credentials.NewTLS(&tls.Config{
		MinVersion: tls.VersionTLS13,
		RootCAs:    cp,
	})
	cc, err := grpc.Dial(rCfg.VideoAPIURL, grpc.WithTransportCredentials(creds))
	if err != nil {
		return fmt.Errorf("cannot create vidi connection: %w", err)
	}
	t.videoapi = videoapi.NewUsersideapiClient(cc)

	// Persist endpoint, since no more errors can be caught
	// until we start making actual requests
	t.state.Endpoint = ep
	if err = t.state.persist(); err != nil {
		return fmt.Errorf("cannot persist state: %w", err)
	}

	t.userapi = userapi.New(&userapi.Config{
		Endpoint: rCfg.UserAPIURL,
	})

	t.logger.Debug("successfully configured ViDi clients",
		zap.String("videoAPI", rCfg.VideoAPIURL),
		zap.String("userAPI", rCfg.UserAPIURL))
	return nil
}

// noClients checks whether clients are initialized and can be used.
func (t *Tool) noClients() bool {
	return t.userapi == nil || t.videoapi == nil
}

// getRemoteConfig retrieves json config from ViDi endpoint.
func (t *Tool) getRemoteConfig(ep string) (*RemoteCFG, error) {
	var rCfg RemoteCFG
	resp, err := t.httpC.NewRequest().
		SetHeader("Accept", "application/json").
		SetResult(&rCfg).Get(ep + configURLPath)
	if err != nil {
		return nil, fmt.Errorf("cannot contact ViDi endpoint: %w", err)
	}
	if resp.IsError() {
		return nil, fmt.Errorf("ViDi endpoint returned err status: %s", resp.Status())
	}
	rCfg.vidiCA, err = base64.StdEncoding.DecodeString(rCfg.VidiCAB64)
	if err != nil {
		return nil, fmt.Errorf("cannot decode vidi ca: %w", err)
	}
	t.logger.Debug("vidi ca decoded", zap.String("vidi_ca", rCfg.VidiCAB64))
	return &rCfg, nil
}
