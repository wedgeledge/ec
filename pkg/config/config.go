package config

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// TODO: add singleton for config.Get()
var config *EdgeConfig

type EdgeProxy struct {
	URL      string `mapstructure:"url"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

type EdgeConfig struct {
	EdgeBaseURL  string `mapstructure:"edge_base_url"`
	EdgeUsername string `mapstructure:"edge_username"`
	EdgePassword string `mapstructure:"edge_password"`
	EdgeProxy    `mapstructure:"edge_proxy"`
}

func Get(confFile string) *EdgeConfig {
	cfg := viper.New()

	// Handle environment variables
	// e.g., EC_EDGE_BASE_URL = edge_base_url = EdgeBaseURL
	cfg.SetEnvPrefix("EC")

	// Explicitly bind known environment variables
	cfg.BindEnv("EDGE_BASE_URL")
	cfg.BindEnv("EDGE_USERNAME")
	cfg.BindEnv("EDGE_PASSWORD")
	cfg.BindEnv("Edge_Proxy.URL", "EDGE_PROXY_URL")
	cfg.BindEnv("Edge_Proxy.Username", "EDGE_PROXY_USERNAME")
	cfg.BindEnv("Edge_Proxy.Password", "EDGE_PROXY_PASSWORD")

	if confFile != "UNSET" {
		cfg.SetConfigFile(confFile)
	} else {
		cfg.SetConfigName("config") // name of config file (without extension)
		cfg.SetConfigType("json")   // REQUIRED if the config file does not have the extension in the name
		cfg.AddConfigPath(".")      // optionally look for config in the working directory
		// TODO: test following paths with .
		cfg.AddConfigPath("$HOME/.config/ec/") // call multiple times to add many search paths
		cfg.AddConfigPath("/etc/ec/")          // call multiple times to add many search paths
	}

	if err := cfg.ReadInConfig(); err != nil {
		// Handle errors
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.WithField("error", err.Error()).Error("Config file error")
		} else {
			log.WithField("error", err.Error()).Error("Config file error")
			os.Exit(1)
		}
	}

	cfg.SetDefault("EdgeBaseURL", "http://localhost:3000")

	/*	if werr := cfg.SafeWriteConfig(); werr != nil {
		fmt.Println("ERROR: writing config: ", werr.Error())
	} */

	err := cfg.Unmarshal(&config)
	if err != nil {
		fmt.Println("ERROR unmarshaling config: ", err.Error())
		os.Exit(1)
	}

	//fmt.Println(*config)

	return config
}
