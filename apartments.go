package main

import (
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/sp0x/rented/sites"
	"github.com/sp0x/torrentd/indexer"
	"github.com/sp0x/torrentd/indexer/categories"
	"github.com/sp0x/torrentd/indexer/definitions"
	"github.com/sp0x/torrentd/indexer/search"
	"github.com/sp0x/torrentd/storage"
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
	chatBroadcastChan := make(chan search.ExternalResultItem)
	go runBot(chatBroadcastChan)
	for true {
		select {
		case result := <-resultsChan:
			//log.Infof("New result: %s\n", result)
			chatBroadcastChan <- result
			if result.IsNew() || result.IsUpdate() {
				price := result.GetField("price")
				reserved := result.GetField("reserved")
				area := result.Size
				fmt.Printf("[%s][%d][%s] %s - %s\n", price, area, reserved, result.ResultItem.Title, result.Link)
			}
		}
	}

	//We store them here also, so we have faster access
	//bolts := storage.BoltStorage{}
	//_ = bolts.StoreSearchResults(currentSearch.GetResults())
	//for _, r := range currentSearch.GetResults() {
	//
	//}
}

//Run the telegram bot.
func runBot(itemsChannel <-chan search.ExternalResultItem) {
	token := viper.GetString("telegram_token")
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		fmt.Printf("Couldn't initialize telegram bot.")
		os.Exit(1)
	}
	//bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)
	bolts, _ := storage.NewBoltStorage("")

	//Listen for people connecting to us
	go func() {
		for update := range updates {
			if update.Message == nil { // ignore any non-Message Updates
				continue
			}
			//We create our chat
			_ = bolts.StoreChat(&storage.Chat{
				Username:    update.Message.From.UserName,
				InitialText: update.Message.Text,
				ChatId:      update.Message.Chat.ID,
			})
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
			//reply := update.Message.Text
			if update.Message.Text == "/start" {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Hello. I'll keep you posted for new apartments.")
				//msg.ReplyToMessageID = update.Message.MessageID
				_, _ = bot.Send(msg)
			}
		}
	}()
	for item := range itemsChannel {
		if !item.IsNew() && !item.IsUpdate() {
			continue
		}
		price := item.GetField("price")
		reserved := item.GetField("reserved")
		if reserved == "true" {
			reserved = "It's currently reserved"
		} else {
			reserved = "And not reserved yet!!!"
		}
		_ = bolts.ForChat(func(chat *storage.Chat) {
			msgText := fmt.Sprintf("I found a new property\n"+
				"[%s](%s)\n"+
				"*%s* - %s", item.Title, item.Link, price, reserved)
			msg := tgbotapi.NewMessage(chat.ChatId, msgText)
			msg.DisableWebPagePreview = false
			msg.ParseMode = "markdown"
			//Since we're not replying.
			//msg.ReplyToMessageID = update.Message.MessageID
			_, _ = bot.Send(msg)
			imgMsg := tgbotapi.NewPhotoUpload(chat.ChatId, nil)
			imgMsg.FileID = item.Banner
			imgMsg.UseExisting = true
			_, _ = bot.Send(imgMsg)
		})
	}

}
