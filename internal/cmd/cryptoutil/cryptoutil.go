// Package cryptoutil contains reusable utility functions
package cryptoutil

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"math"
	"math/big"
	"net/url"
	"time"
)

// GenerateTestCerts generates a root for the trust domain and uses it to sign an x509 SVID for each supplied SPIFFEID
func GenerateTestCerts(spiffeids ...string) ([]tls.Certificate, error) {
	var tlsCerts []tls.Certificate

	// if no spiffe IDs were requested, then we can't return anything since we don't know what the trust domain is going to be
	if len(spiffeids) == 0 {
		return []tls.Certificate{}, nil
	}

	// assume that all spiffe IDs have the same trustdomain, use the Host of the first SPIFFEID to set the uri for the CA
	uri, err := url.Parse(spiffeids[0])
	if err != nil {
		return []tls.Certificate{}, err
	}
	caUri, err := url.Parse("spiffe://" + uri.Host)
	if err != nil {
		return []tls.Certificate{}, err
	}

	// create the CA cert
	caKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return []tls.Certificate{}, err
	}

	caSerial, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
	if err != nil {
		return []tls.Certificate{}, err
	}
	caSubj := pkix.Name{
		Country:            []string{"GB"},
		Organization:       []string{"Jetstack"},
		OrganizationalUnit: []string{"Product"},
		SerialNumber:       caSerial.String(),
	}
	caTemplate := &x509.Certificate{
		BasicConstraintsValid: true,
		SignatureAlgorithm:    x509.ECDSAWithSHA256,
		PublicKeyAlgorithm:    x509.ECDSA,
		PublicKey:             caKey.Public(),
		SerialNumber:          caSerial,
		Issuer:                caSubj,
		Subject:               caSubj,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(100 * time.Hour * 24 * 365),
		KeyUsage:              x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IsCA:                  true,
		DNSNames:              nil,
		EmailAddresses:        nil,
		IPAddresses:           nil,
		URIs:                  []*url.URL{caUri},
	}

	caCert, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, caKey.Public(), caKey)
	if err != nil {
		return []tls.Certificate{}, err
	}

	// create certs for each of the requested spiffe IDS
	for _, spiffeID := range spiffeids {
		var tlsCert tls.Certificate

		leafKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			return []tls.Certificate{}, err
		}
		tlsCert.PrivateKey = leafKey

		leafSerial, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
		if err != nil {
			return []tls.Certificate{}, err
		}
		leafSubj := pkix.Name{
			Country:            []string{"GB"},
			Organization:       []string{"Jetstack"},
			OrganizationalUnit: []string{"Product"},
			SerialNumber:       leafSerial.String(),
		}
		uri, err := url.Parse(spiffeID)
		if err != nil {
			return []tls.Certificate{}, err
		}

		leafTemplate := &x509.Certificate{
			BasicConstraintsValid: true,
			SignatureAlgorithm:    x509.ECDSAWithSHA256,
			PublicKeyAlgorithm:    x509.ECDSA,
			PublicKey:             leafKey.Public(),
			SerialNumber:          leafSerial,
			Issuer:                caSubj,
			Subject:               leafSubj,
			NotBefore:             time.Now(),
			NotAfter:              time.Now().Add(99 * time.Hour * 24 * 365),
			KeyUsage:              x509.KeyUsageDigitalSignature,
			ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
			IsCA:                  false,
			DNSNames:              nil,
			EmailAddresses:        nil,
			IPAddresses:           nil,
			URIs:                  []*url.URL{uri},
		}

		leafCert, err := x509.CreateCertificate(rand.Reader, leafTemplate, caTemplate, leafKey.Public(), caKey)
		if err != nil {
			return []tls.Certificate{}, err
		}
		tlsCert.Certificate = [][]byte{leafCert, caCert}

		tlsCerts = append(tlsCerts, tlsCert)
	}

	return tlsCerts, nil
}
