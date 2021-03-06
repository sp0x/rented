package main

import (
	"fmt"
	"github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/sp0x/torrentd/config"
	"github.com/spf13/viper"
	"os"
	"path"
)

func initConfig(appName string) config.ViperConfig {
	//We load the default config file
	homeDir, _ := homedir.Dir()
	if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		defaultConfigPath := path.Join(homeDir, fmt.Sprintf(".%s", appName))
		_ = os.MkdirAll(defaultConfigPath, os.ModePerm)
		viper.AddConfigPath(defaultConfigPath)
		viper.SetConfigType("yaml")
		viper.SetConfigName(fmt.Sprintf(".%s", appName))
	}
	viper.AutomaticEnv()
	var appConfig config.ViperConfig
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			config.SetDefaults(&appConfig)
			err = viper.SafeWriteConfig()
			if err != nil {
				log.Warningf("error while writing default config file: %v\n", err)
			}
		} else {
			log.Warningf("error while reading config file: %v\n", err)
		}
	}
	if viper.GetBool("verbose") {
		log.SetLevel(log.DebugLevel)
	} else {

	}
	return appConfig

}
