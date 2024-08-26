package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/go-co-op/gocron"
)

// fatalError сообщает об ошибке и выходит из приложения
func fatalError(e error) {
	slog.Error("Фатальная ошибка. Выход из приложения", slog.String("error", e.Error()))
	os.Exit(1)
}

func main() {
	// Инициализация логирования
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	// Инициализация конфигурации
	initConfig()

	// Подготовка источников и базы данных
	if err := prepare(); err != nil {
		fatalError(err)
	}

	// Запуск задач для работы с файлами
	scheduler := gocron.NewScheduler(time.Local)
	jobReadZipFiles, _ := scheduler.Every(60).Second().Do(parseZipFiles)
	jobReadZipFiles.SingletonMode()
	scheduler.StartBlocking()
}
