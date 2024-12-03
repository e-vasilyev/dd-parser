package main

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

var config = viper.NewWithOptions(viper.EnvKeyReplacer(strings.NewReplacer(".", "_")))

func initConfig() {
	// Настройки переменных среды
	config.SetEnvPrefix("DDP")
	config.AutomaticEnv()

	// Настройки базы данных
	config.SetDefault("database.name", "dd_parser")
	config.SetDefault("database.host", "localhost")
	config.SetDefault("database.port", "5432")
	config.SetDefault("database.username", "postgres")
	config.SetDefault("database.password", "postgres")
	config.Set("database.url", fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		config.GetString("database.username"),
		config.GetString("database.password"),
		config.GetString("database.host"),
		config.GetString("database.port"),
		config.GetString("database.name"),
	))

	// Настройки источников
	config.SetDefault("source.type", "local")

	// Настройки локального источника
	config.SetDefault("source.local.root_path", "source")

	// Настройки s3
	config.SetDefault("source.s3.bucket_name", "diadoc")
	config.SetDefault("source.s3.endpoint", "localhost:9000")
	config.SetDefault("source.s3.password", "password")
	config.SetDefault("source.s3.user", "root")
	config.SetDefault("source.s3.use_ssl", false)
	config.SetDefault("source.s3.use_root", true)
}
