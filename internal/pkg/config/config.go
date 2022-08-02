package config

import (
	"fmt"
	"io/fs"
	"sync/atomic"

	"gopkg.in/yaml.v2"

	"github.com/jetstack/spiffe-demo/types"
)

var (
	currentConfig atomic.Value // *types.ConfigFile
	currentSource atomic.Value // *SpiffeDemoSource

	CurrentSource DynamicSource
)

func init() {
	currentConfig.Store(new(types.ConfigFile))
	currentSource.Store(new(SpiffeDemoSource))
}

func ReadConfigFromFS(fsys fs.FS, path string) (*types.ConfigFile, error) {
	rawConfig, err := fs.ReadFile(fsys, path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %s", err)
	}

	var cfg types.ConfigFile

	err = yaml.Unmarshal(rawConfig, &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %s", err)
	}

	return &cfg, nil
}

func ReadAndStoreConfig(fsys fs.FS, path string) error {
	config, err := ReadConfigFromFS(fsys, path)
	if err != nil {
		return err
	}
	StoreConfig(config)
	return nil
}

func StoreConfig(cfg *types.ConfigFile) {
	currentConfig.Store(cfg)
}

func GetCurrentConfig() *types.ConfigFile {
	return currentConfig.Load().(*types.ConfigFile)
}

func StoreCurrentSource(source *SpiffeDemoSource) {
	currentSource.Store(source)
}

func GetCurrentSource() *SpiffeDemoSource {
	return currentSource.Load().(*SpiffeDemoSource)
}
