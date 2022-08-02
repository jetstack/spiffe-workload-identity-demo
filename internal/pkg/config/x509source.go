package config

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync/atomic"

	"github.com/spiffe/go-spiffe/v2/bundle/x509bundle"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/spiffe/go-spiffe/v2/svid/x509svid"
	"github.com/spiffe/go-spiffe/v2/workloadapi"

	"github.com/jetstack/spiffe-demo/types"
)

// Interface guards
var _ x509svid.Source = &SpiffeDemoSource{}
var _ x509bundle.Source = &SpiffeDemoSource{}

// SpiffeDemoSource implements x509svid.Source and x509bundle.Source by either
// reading files or communicating with the SPIRE workload API.
type SpiffeDemoSource struct {
	cancelFunc context.CancelFunc

	workloadAPISource *workloadapi.X509Source

	currentSVID        atomic.Value // *x509svid.SVID
	currentTrustBundle atomic.Value // *x509bundle.Bundle
}

// ConstructSpiffeDemoSource constructs a new SPIFFE Connector source ready to become the current source.
// When disposing of the source be sure to cancel the Context, as this will clean up the fsnotify watchers.
func ConstructSpiffeDemoSource(ctx context.Context, cancel context.CancelFunc, config *types.SpiffeConfig) (*SpiffeDemoSource, error) {
	source := &SpiffeDemoSource{
		cancelFunc: cancel,
	}
	if config == nil {
		return nil, errors.New("no SPIFFE config provided")
	}

	// If Workload API is set, just use that.
	if config.SVIDSources.WorkloadAPI != nil {
		x509source, err := workloadapi.NewX509Source(ctx, workloadapi.WithClientOptions(workloadapi.WithAddr(config.SVIDSources.WorkloadAPI.SocketPath)))
		if err != nil {
			return nil, err
		}
		source.workloadAPISource = x509source
		return source, nil
	}

	if config.SVIDSources.InMemory != nil {
		svid, err := x509svid.Parse(config.SVIDSources.InMemory.SVIDCert, config.SVIDSources.InMemory.SVIDKey)
		if err != nil {
			return source, err
		}
		if svid == nil {
			return source, errors.New("no SVID provided in config file")
		}
		source.currentSVID.Store(svid)

		bundle, err := x509bundle.Parse(spiffeid.RequireTrustDomainFromString("todo"), config.SVIDSources.InMemory.TrustDomainCA)
		if err != nil {
			return source, err
		}
		source.currentTrustBundle.Store(bundle)

		return source, nil
	}

	// Otherwise, start watching files for SVIDs and Trust bundles.
	if config.SVIDSources.Files == nil {
		return nil, errors.New("neither workload API nor files provided in config file")
	}
	if _, err := os.Stat(config.SVIDSources.Files.SVIDKey); err != nil {
		return nil, fmt.Errorf("could not read SVID Key file (%w)", err)
	}

	source.currentSVID.Store(new(x509svid.SVID))
	source.currentTrustBundle.Store(new(x509bundle.Bundle))

	// Start watching for SVID updates
	updateSVID := func() error {
		if config.SVIDSources.Files == nil {
			return errors.New("no SVID Sources speficied in config file")
		}
		svid, err := x509svid.Load(config.SVIDSources.Files.SVIDCert, config.SVIDSources.Files.SVIDKey)
		if err != nil {
			return fmt.Errorf("failed to load SVID: %w", err)
		}
		if svid == nil {
			return errors.New("no SVID provided in config file")
		}
		source.currentSVID.Store(svid)
		return nil
	}
	if err := updateSVID(); err != nil {
		return nil, err
	}
	if _, err := NewWatcher(ctx, config.SVIDSources.Files.SVIDKey, updateSVID); err != nil {
		return nil, err
	}

	// Start watching for Trust bundle updates
	updateTrustBundle := func() error {
		bundle, err := x509bundle.Load(spiffeid.RequireTrustDomainFromString("todo"), config.SVIDSources.Files.TrustDomainCA)
		if err != nil {
			return fmt.Errorf("failed to load trust bundle: %w", err)
		}
		source.currentTrustBundle.Store(bundle)
		return nil
	}
	if err := updateTrustBundle(); err != nil {
		return nil, err
	}
	if _, err := NewWatcher(ctx, config.SVIDSources.Files.TrustDomainCA, updateTrustBundle); err != nil {
		return nil, fmt.Errorf("failed to start new config watcher: %w", err)
	}
	return source, nil
}

func (s *SpiffeDemoSource) GetX509SVID() (*x509svid.SVID, error) {
	if s.workloadAPISource != nil {
		return s.workloadAPISource.GetX509SVID()
	}
	return s.currentSVID.Load().(*x509svid.SVID), nil
}

func (s *SpiffeDemoSource) GetX509BundleForTrustDomain(trustDomain spiffeid.TrustDomain) (*x509bundle.Bundle, error) {
	if s.workloadAPISource != nil {
		return s.workloadAPISource.GetX509BundleForTrustDomain(trustDomain)
	}
	return s.currentTrustBundle.Load().(*x509bundle.Bundle), nil
}

func (s *SpiffeDemoSource) Cancel() {
	s.cancelFunc()
}

// DynamicSource represents the most up-to-date SVID / Trust bundle we have
// from the most recently loaded source config file
type DynamicSource struct{}

func (d DynamicSource) GetX509SVID() (*x509svid.SVID, error) {
	return GetCurrentSource().GetX509SVID()
}

func (d DynamicSource) GetX509BundleForTrustDomain(trustDomain spiffeid.TrustDomain) (*x509bundle.Bundle, error) {
	return GetCurrentSource().GetX509BundleForTrustDomain(trustDomain)
}
