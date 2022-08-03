// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package southbound

import (
	"crypto/tls"
	"fmt"
	topoapi "github.com/onosproject/onos-api/go/onos/topo"
	"github.com/onosproject/onos-lib-go/pkg/certs"
	"github.com/onosproject/onos-lib-go/pkg/errors"
	"time"
)

const (
	defaultTimeout = 60 * time.Second
)

// Destination contains data used to connect to a server
type Destination struct {
	// Addrs is a slice of addresses by which a target may be reached. Most
	// clients will only handle the first element.
	Addrs []string
	// TargetID is the topology target entity ID
	TargetID string
	// Timeout is the connection timeout
	Timeout time.Duration
	// TLS config to use when connecting to target.
	TLS *tls.Config
	// DeviceID is the numerical ID to be used for p4runtime API interactions
	DeviceID uint64
}

func newDestination(target *topoapi.Object) (*Destination, error) {
	p4rtServerInfo := &topoapi.P4RTServerInfo{}
	err := target.GetAspect(p4rtServerInfo)
	if err != nil {
		return nil, errors.NewInvalid("target entity %s must have 'onos.topo.Configurable' aspect to work with controller", target.ID)
	}

	tlsOptions := &topoapi.TLSOptions{}
	err = target.GetAspect(tlsOptions)
	if err != nil {
		return nil, errors.NewInvalid("topo entity %s must have 'onos.topo.TLSOptions' aspect to work with controller", target.ID)
	}

	timeout := defaultTimeout
	if p4rtServerInfo.Timeout != nil {
		timeout = *p4rtServerInfo.Timeout
	}

	p4rtServerAddr := p4rtServerInfo.GetControlEndpoint().GetAddress()
	p4rtServerPort := p4rtServerInfo.GetControlEndpoint().GetPort()
	addr := fmt.Sprintf("%s:%d", p4rtServerAddr, p4rtServerPort)
	destination := &Destination{
		Addrs:    []string{addr},
		TargetID: string(target.ID),
		Timeout:  timeout,
	}

	if tlsOptions.Plain {
		log.Infow("Plain (non TLS) connection to", "P4 runtime server address", p4rtServerAddr)
	} else {
		tlsConfig := &tls.Config{}
		if tlsOptions.Insecure {
			log.Infow("Insecure TLS connection to ", "P4 runtime server address", p4rtServerAddr)
			tlsConfig = &tls.Config{InsecureSkipVerify: true}
		} else {
			log.Infow("Secure TLS connection to ", "P4 runtime server address", p4rtServerAddr)
		}
		if tlsOptions.CaCert == "" {
			log.Info("Loading default CA")
			defaultCertPool, err := certs.GetCertPoolDefault()
			if err != nil {
				return nil, err
			}
			tlsConfig.RootCAs = defaultCertPool
		} else {
			certPool, err := certs.GetCertPool(tlsOptions.CaCert)
			if err != nil {
				return nil, err
			}
			tlsConfig.RootCAs = certPool
		}
		if tlsOptions.Cert == "" && tlsOptions.Key == "" {
			log.Info("Loading default certificates")
			clientCerts, err := tls.X509KeyPair([]byte(certs.DefaultClientCrt), []byte(certs.DefaultClientKey))
			if err != nil {
				return nil, err
			}
			tlsConfig.Certificates = []tls.Certificate{clientCerts}
		} else if tlsOptions.Cert != "" && tlsOptions.Key != "" {
			// Load certs given for device
			tlsCerts, err := setCertificate(tlsOptions.Cert, tlsOptions.Key)
			if err != nil {
				return nil, errors.NewInvalid(err.Error())
			}
			tlsConfig.Certificates = []tls.Certificate{tlsCerts}
		} else {
			log.Errorw("Can't load Ca=%s , Cert=%s , key=%s for %v, trying with insecure connection",
				"CA", tlsOptions.CaCert, "Cert", tlsOptions.Cert, "Key", tlsOptions.Key, "P4 runtime server address", p4rtServerAddr)
			tlsConfig = &tls.Config{InsecureSkipVerify: true}
		}
		destination.TLS = tlsConfig
	}

	err = destination.Validate()
	if err != nil {
		return nil, err
	}

	return destination, nil
}

// Validate validates the fields of Destination.
func (d Destination) Validate() error {
	// TODO add validate checks
	return nil
}

func setCertificate(pathCert string, pathKey string) (tls.Certificate, error) {
	certificate, err := tls.LoadX509KeyPair(pathCert, pathKey)
	if err != nil {
		return tls.Certificate{}, errors.NewNotFound("could not load client key pair ", err)
	}
	return certificate, nil
}
