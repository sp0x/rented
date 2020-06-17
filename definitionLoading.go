package main

import (
	"fmt"
	"github.com/sp0x/rented/sites"
	"github.com/sp0x/torrentd/indexer"
	"github.com/sp0x/torrentd/indexer/definitions"
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

func getIndexLoader() *indexer.MultipleDefinitionLoader {
	return &indexer.MultipleDefinitionLoader{
		getEmbeddedDefinitionsSource(),
		indexer.NewFsLoader(appName),
	}
}
