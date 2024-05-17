//go:build e2e

package e2e

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	mrand "math/rand"
	"net"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/adwski/vidi/internal/app/processor"
	"github.com/adwski/vidi/internal/app/streamer"
	"github.com/adwski/vidi/internal/app/uploader"
	"github.com/adwski/vidi/internal/app/user"
	"github.com/adwski/vidi/internal/app/video"
)

func TestMain(m *testing.M) {
	if err := generateCertAndKey("localhost"); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var (
		wg          = &sync.WaitGroup{}
		ctx, cancel = context.WithCancel(context.Background())
	)

	wg.Add(5)
	go func() {
		user.NewApp().RunWithContextAndConfig(ctx, "userapi.yaml")
		wg.Done()
	}()
	go func() {
		video.NewApp().RunWithContextAndConfig(ctx, "videoapi.yaml")
		wg.Done()
	}()
	go func() {
		uploader.NewApp().RunWithContextAndConfig(ctx, "uploader.yaml")
		wg.Done()
	}()
	go func() {
		processor.NewApp().RunWithContextAndConfig(ctx, "processor.yaml")
		wg.Done()
	}()
	go func() {
		streamer.NewApp().RunWithContextAndConfig(ctx, "streamer.yaml")
		wg.Done()
	}()

	time.Sleep(5 * time.Second)

	code := m.Run()
	cancel()
	wg.Wait()
	defer func() {
		os.Exit(code)
	}()
}

const (
	caOrg        = "VIDItest"
	caCountry    = "RU"
	caValidYears = 10

	privateKeyRSALen = 4096

	certPath = "cert.pem"
	keyPath  = "key.pem"
)

var (
	caSubjectKeyIdentifier = []byte{1, 2, 3, 4, 6}
)

func generateCertAndKey(cn string) error {
	key, err := rsa.GenerateKey(rand.Reader, privateKeyRSALen)
	if err != nil {
		return fmt.Errorf("cannot generate rsa private key: %w", err)
	}
	ca := getCA(cn)
	cert, err := x509.CreateCertificate(rand.Reader, ca, ca, &key.PublicKey, key)
	if err != nil {
		return fmt.Errorf("cannot create x509 certificate: %w", err)
	}
	keyPem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(key),
		},
	)
	certPem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: cert,
		},
	)
	if err = os.WriteFile(keyPath, keyPem, 0600); err != nil {
		return fmt.Errorf("cannot write private key to file: %w", err)
	}
	if err = os.WriteFile(certPath, certPem, 0600); err != nil {
		return fmt.Errorf("cannot write cert to file: %w", err)
	}
	return nil
}

func getCA(cn string) *x509.Certificate {
	return &x509.Certificate{
		SerialNumber: big.NewInt(mrand.Int63()),
		Subject: pkix.Name{
			Country:      []string{caCountry},
			Organization: []string{caOrg},
			CommonName:   cn,
		},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback}, //nolint:gomnd // ip addr
		DNSNames:     []string{cn},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(caValidYears, 0, 0),
		SubjectKeyId: caSubjectKeyIdentifier,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}
}
