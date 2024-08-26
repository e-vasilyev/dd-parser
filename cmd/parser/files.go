package main

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"

	"golang.org/x/text/encoding/charmap"
)

// createLocalDir создает локальную директорию
func createLocalDir(path string) error {
	return os.MkdirAll(path, os.ModePerm)
}

// prepareLocalDir подгатавливает локальные директории
func prepareLocalDirs() error {
	slog.Info("Выбран локальный тип источника")
	rootPath := config.GetString("source.local.rootPath")
	var paths = [3]string{"zip", "error", "done"}
	for _, path := range paths {
		if err := createLocalDir(filepath.Join(rootPath, path)); err != nil {
			return err
		}
	}
	absRootPath, err := filepath.Abs(rootPath)
	if err != nil {
		return err
	}
	slog.Info(fmt.Sprintf("Корневая директория источника: %s", absRootPath))
	return nil
}

// manageFilesInZip обрабатывает файлы в zip архиве
func manageFilesInZip(z *zip.ReadCloser) {
	for _, file := range z.File {
		var diadoc = &DiaDocXML{}
		slog.Info(fmt.Sprintf("Обработка файла %s", file.Name))
		openFile, _ := file.Open()
		decoder := xml.NewDecoder(openFile)
		decoder.CharsetReader = func(charset string, input io.Reader) (io.Reader, error) {
			switch charset {
			case "windows-1251":
				return charmap.Windows1251.NewDecoder().Reader(input), nil
			default:
				return nil, fmt.Errorf("не поддерживаемая кодировка: %s", charset)
			}
		}
		err := decoder.Decode(&diadoc)
		if err != nil {
			slog.Error("Ошибка разбора xml", slog.String("error", err.Error()))
		}
	}
}

// parseLocalZipFiles читает и распаковывает найденные zip файлы из локальной директории
func parseLocalZipFiles() error {
	return filepath.Walk(filepath.Join(config.GetString("source.local.rootPath"), "zip"), func(path string, info fs.FileInfo, err error) error {
		if err == nil && filepath.Ext(path) == ".zip" {
			slog.Info(fmt.Sprintf("Найден архив %s", info.Name()))
			files, err := zip.OpenReader(path)
			if err == nil {
				slog.Debug("Обработка архива")
				defer files.Close()
				manageFilesInZip(files)
			} else {
				slog.Error("Ошибка при чтении архива", slog.String("error", err.Error()))
			}
		}
		return err
	})
}

// parseZipFiles обрабатывает найденные zip файлы
func parseZipFiles() error {
	var err error
	slog.Debug("Проверка наличия файлов")
	switch config.GetString("source.type") {
	case "local":
		err = parseLocalZipFiles()
	default:
		return fmt.Errorf("не поддерживаемый тип источника %s", config.GetString("source.type"))
	}
	return err
}

// prepare подгатавливает приложение к работе с источниками и базами данных
func prepare() error {
	var err error
	slog.Debug("Определение типа источнка")
	switch config.GetString("source.type") {
	case "local":
		err = prepareLocalDirs()
	default:
		return fmt.Errorf("не поддерживаемый тип источника %s", config.GetString("source.type"))
	}
	return err
}
