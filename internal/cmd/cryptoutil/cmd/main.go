package main

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"

	"github.com/jetstack/spiffe-demo/internal/cmd/cryptoutil"
)

const testConfig = `---
spiffe:
  svid_sources:
    files:
      trust_domain_ca: ./ca.pem
      svid_cert: ./svid_1_cert.pem
      svid_key: ./svid_1_key.pem
`

// Generate testing material for use locally.
func main() {
	if len(os.Args) < 2 {
		exitOnErr(errors.New("usage: " + os.Args[0] + " 'spiffe://your.domain/your/id' [...]"))
	}

	certs, err := cryptoutil.GenerateTestCerts(os.Args[1:len(os.Args)]...)
	exitOnErr(err)

	caCert := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certs[0].Certificate[1],
	})
	exitOnErr(os.WriteFile("ca.pem", caCert, 0o600))

	for i, c := range certs {
		leafCert := pem.EncodeToMemory(&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: c.Certificate[0],
		})

		x509Key, _ := x509.MarshalPKCS8PrivateKey(c.PrivateKey)
		key := pem.EncodeToMemory(&pem.Block{
			Type:  "PRIVATE KEY",
			Bytes: x509Key,
		})

		exitOnErr(os.WriteFile(fmt.Sprintf("svid_%d_cert.pem", i+1), leafCert, 0o600))
		exitOnErr(os.WriteFile(fmt.Sprintf("svid_%d_key.pem", i+1), key, 0o666))
	}

	// test config will only contain the first SVID
	exitOnErr(os.WriteFile("test.yaml", []byte(testConfig), 0o666))
}

func exitOnErr(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't generate certs: %s\n", err.Error())
		os.Exit(1)
	}
}
