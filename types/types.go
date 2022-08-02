// Package types contains the config file structs
package types

// ConfigFile represents the config file that will be loaded from disk, or some other mechanism.
type ConfigFile struct {
	SPIFFE *SpiffeConfig `yaml:"spiffe"`
}

// SpiffeConfig represents the SPIFFE configuration section of spiffe-connector's config file
type SpiffeConfig struct {
	SVIDSources SVIDSources `yaml:"svid_sources"`
}

// SVIDSources determines where spiffe-connector will obtain its own SVID and trust domain information.
// The SPIFFE Workload API and Static files are supported.
type SVIDSources struct {
	WorkloadAPI *WorkloadAPI `yaml:"workload_api,omitempty"`
	Files       *Files       `yaml:"files,omitempty"`

	// InMemory is only used in testing
	InMemory *InMemory
}

type WorkloadAPI struct {
	SocketPath string `yaml:"socket_path"`
}

type Files struct {
	TrustDomainCA string `yaml:"trust_domain_ca"`
	SVIDCert      string `yaml:"svid_cert"`
	SVIDKey       string `yaml:"svid_key"`
}

// InMemory is only used in testing
type InMemory struct {
	TrustDomainCA []byte
	SVIDCert      []byte
	SVIDKey       []byte
}
