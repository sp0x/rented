package main

import (
	"fmt"
	"github.com/mitchellh/go-homedir"
	"github.com/sp0x/rented/sites"
	"github.com/sp0x/torrentd/indexer"
	"github.com/sp0x/torrentd/indexer/definitions"
	"os"
	"path"
	"path/filepath"
)

func getEmbeddedDefinitionsSource() indexer.DefinitionLoader {
	x := indexer.CreateEmbeddedDefinitionSource(sites.GzipAssetNames(), func(key string) ([]byte, error) {
		fullname := fmt.Sprintf("sites/%s.yml", key)
		data, err := definitions.GzipAsset(fullname)
		if err != nil {
			return nil, err
		}
		data, _ = definitions.UnzipData(data)
		return data, nil
	})
	return x
}

func getFileDefinitionsSource() indexer.DefinitionLoader {
	localDirectory := ""
	if cwd, err := os.Getwd(); err == nil {
		localDirectory = filepath.Join(cwd, "sites")
	}
	home, _ := homedir.Dir()
	homeDefsDir := path.Join(home, ".rented", "sites")
	x := &indexer.FileIndexLoader{
		Directories: []string{
			localDirectory,
			homeDefsDir,
		},
	}
	return x
}

func getIndexLoader() *indexer.MultipleDefinitionLoader {
	return &indexer.MultipleDefinitionLoader{
		getEmbeddedDefinitionsSource(),
		getFileDefinitionsSource(),
	}
}
