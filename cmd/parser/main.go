package main

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/go-co-op/gocron"
)

var (
	logLevelString string = os.Getenv("DDP_LOG_LEVEL")
	pdb            *PDB   = &PDB{pool: nil}
	s3             *S3    = &S3{client: nil}
)

func main() {
	// Инициализация логирования
	logLevel := slog.LevelInfo
	switch strings.ToLower(logLevelString) {
	case "debug":
		logLevel = slog.LevelDebug
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
	slog.SetDefault(logger)

	// Инициализация конфигурации
	initConfig()

	// Подготовка источников
	if err := prepareSource(); err != nil {
		slog.Error(fmt.Sprintf("Критическая ошибка при подготовке источника данных: %s", err.Error()))
		os.Exit(1)
	}

	// Подключение к БД
	pool, err := connectToDB()
	if err != nil {
		slog.Error(fmt.Sprintf("Критическая ошибка при подключении к базе данных: %s", err.Error()))
		os.Exit(1)
	}
	pdb.pool = pool
	defer pdb.pool.Close()

	// Миграция базы данных
	if err := pdb.migration(); err != nil {
		slog.Error(fmt.Sprintf("Критическая ошибка при подключении миграции базы данных: %s", err.Error()))
		os.Exit(1)
	}

	// Запуск задач для работы с файлами
	scheduler := gocron.NewScheduler(time.Local)
	jobReadZipFiles, _ := scheduler.Every(60).Second().Do(parseZipFiles)
	jobReadZipFiles.SingletonMode()
	scheduler.StartBlocking()
}
