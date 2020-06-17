package main

import (
	"fmt"
	"github.com/sp0x/torrentd/config"
	"github.com/sp0x/torrentd/db"
	"github.com/sp0x/torrentd/indexer/search"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

var appName = "rented"
var rootCmd = &cobra.Command{
	Use:   appName,
	Short: "Apartment tracker service.",
	Run:   findApartments,
}
var configFile = ""
var appConfig config.ViperConfig

func migrateDb() {
	_ = os.MkdirAll("./db", os.ModePerm)
	gormDb := db.GetOrmDb("")
	defer gormDb.Close()
	gormDb.AutoMigrate(&search.ExternalResultItem{})
}

func init() {
	//Init our db
	migrateDb()
	cobra.OnInitialize(func() {
		appConfig = initConfig(appName)
	})
	flags := rootCmd.PersistentFlags()
	var verbose bool
	flags.BoolVarP(&verbose, "verbose", "v", false, "Show more logs")
	flags.StringVar(&configFile, "config", "", fmt.Sprintf("The configuration file to use. By default it is ~/.%s/.%s.yaml",
		appName, appName))
	_ = viper.BindPFlag("verbose", flags.Lookup("verbose"))
	_ = viper.BindEnv("verbose")

	localFlags := rootCmd.Flags()
	localFlags.StringVarP(&aptIndexer, "indexer", "x", "cityapartment", "The apartment site to use.")
	_ = viper.BindPFlag("indexer", flags.Lookup("indexer"))
}

func main() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
