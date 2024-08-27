package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/go-co-op/gocron"
)

var pdb *PDB = &PDB{pool: nil}

func main() {
	// Инициализация логирования
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	// Инициализация конфигурации
	initConfig()

	// Подготовка источников
	if err := prepareSource(); err != nil {
		slog.Error("Ошибка при подготовке источника данных. Выход из приложения", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Подключение к БД
	pool, err := connectToDB()
	if err != nil {
		slog.Error("Ошибка при подключении к базе данных. Выход из приложения", slog.String("error", err.Error()))
		os.Exit(1)
	}
	pdb.pool = pool
	defer pdb.pool.Close()

	// Миграция базы данных
	if err := pdb.migration(); err != nil {
		slog.Error("Ошибка при подключении миграции базы данных. Выход из приложения", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Запуск задач для работы с файлами
	scheduler := gocron.NewScheduler(time.Local)
	jobReadZipFiles, _ := scheduler.Every(60).Second().Do(parseZipFiles)
	jobReadZipFiles.SingletonMode()
	scheduler.StartBlocking()
}
