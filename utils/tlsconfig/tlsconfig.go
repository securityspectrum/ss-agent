package tlsconfig

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"

	"ss-agent/config"
)

func SetupTLSConfig() (*tls.Config, error) {
	conf := config.GetConfig()
	cert, err := tls.LoadX509KeyPair(conf.CertFile, conf.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load client certificate: %v", err)
	}

	caCert, err := ioutil.ReadFile(conf.CAFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA certificate: %v", err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
	}, nil
}
