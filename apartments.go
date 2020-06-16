package main

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/mitchellh/go-homedir"
	"github.com/sp0x/rented/bots"
	"github.com/sp0x/rented/sites"
	"github.com/sp0x/torrentd/indexer"
	"github.com/sp0x/torrentd/indexer/categories"
	"github.com/sp0x/torrentd/indexer/definitions"
	"github.com/sp0x/torrentd/indexer/search"
	"github.com/sp0x/torrentd/torznab"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"path"
	"path/filepath"
	"strings"
)

var aptIndexer string

func init() {
	//flags := cmdGetApartments.Flags()
	_ = viper.BindEnv("indexer")
	_ = viper.BindEnv("telegram_token")
}

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

func findApartments(cmd *cobra.Command, args []string) {
	indexer.Loader = getIndexLoader()

	//Construct our facade based on the needed indexer.
	indexerFacade, err := indexer.NewFacade(aptIndexer, &appConfig, categories.Rental)
	if err != nil {
		fmt.Printf("Couldn't initialize the named indexer `%s`: %s", aptIndexer, err)
		os.Exit(1)
	}
	if indexerFacade == nil {
		fmt.Printf("Indexer facade was nil")
		os.Exit(1)
	}
	var searchQuery = strings.Join(args, " ")
	watchIntervalSec := 30
	query := torznab.ParseQueryString(searchQuery)
	query.AddCategory(categories.Rental)
	resultsChan := indexer.Watch(indexerFacade, query, watchIntervalSec)
	readIndexer(resultsChan)
}

//Reads the channel that's the result of watching an indexer.
func readIndexer(resultsChan <-chan search.ExternalResultItem) {
	chatMessagesChannel := make(chan bots.ChatMessage)
	token := viper.GetString("TELEGRAM_TOKEN")
	telegram, err := bots.NewTelegram(token, tgbotapi.NewBotAPI)
	if err != nil {
		fmt.Printf("Couldn't initialize telegram bot: %v", err)
		os.Exit(1)
	}
	go func() {
		_ = telegram.Run()
	}()
	go func() {
		_ = telegram.FeedBroadcast(chatMessagesChannel)
	}()
	for {
		result := <-resultsChan
		//log.Infof("New result: %s\n", result)
		if result.IsNew() || result.IsUpdate() {
			price := result.GetField("price")
			reserved := result.GetField("reserved")
			if reserved == "true" {
				reserved = "It's currently reserved"
			} else {
				reserved = "And not reserved yet!!!"
			}
			msgText := fmt.Sprintf("I found a new property\n"+
				"[%s](%s)\n"+
				"*%s* - %s", result.Title, result.Link, price, reserved)
			message := bots.ChatMessage{Text: msgText, Banner: result.Banner}
			chatMessagesChannel <- message
			area := result.Size
			fmt.Printf("[%s][%d][%s] %s - %s\n", price, area, reserved, result.ResultItem.Title, result.Link)
		}

	}

	//We store them here also, so we have faster access
	//bolts := storage.BoltStorage{}
	//_ = bolts.StoreSearchResults(currentSearch.GetResults())
	//for _, r := range currentSearch.GetResults() {
	//
	//}
}
