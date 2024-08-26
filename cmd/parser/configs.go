package main

import (
	"strings"

	"github.com/spf13/viper"
)

var config = viper.NewWithOptions(viper.EnvKeyReplacer(strings.NewReplacer(".", "_")))

func initConfig() {
	config.SetEnvPrefix("DDP_")
	config.AutomaticEnv()

	config.SetDefault("source.type", "local")
	config.SetDefault("source.local.rootPath", "source")
}
