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
	var paths = [3]string{"zip", "error", "done"}

	slog.Info("Выбран локальный тип источника")

	rootPath := config.GetString("source.local.rootPath")
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

// manageFilesInZip обрабатывает файлы в zip архиве.
// Возвращает количество ошибочных документов.
func manageFilesInZip(z *zip.ReadCloser) int {
	var countErr int

	slog.Info(fmt.Sprintf("Найдено %v документов в архиве", len(z.File)))

	for _, file := range z.File {
		document, err := parseXML(file)
		if err != nil {
			slog.Error("Ошибка разбора XML", slog.String("error", err.Error()))
			countErr++
		} else {
			slog.Debug(fmt.Sprintf("Разбор документа %s успешно завершен", file.Name))
			if err := pdb.insertDocument(document); err != nil {
				countErr++
			}
		}
	}

	slog.Info(fmt.Sprintf("Обработано %v из %v документов в архиве", len(z.File)-countErr, len(z.File)))

	return countErr
}

// parseXML разбирает XML
func parseXML(file *zip.File) (*DiaDocXML, error) {
	var diadoc = &DiaDocXML{}

	slog.Info(fmt.Sprintf("Обработка файла %s", file.Name))

	openFile, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer openFile.Close()

	decoder := xml.NewDecoder(openFile)
	decoder.CharsetReader = func(charset string, input io.Reader) (io.Reader, error) {
		switch charset {
		case "windows-1251":
			return charmap.Windows1251.NewDecoder().Reader(input), nil
		default:
			return nil, fmt.Errorf("не поддерживаемая кодировка: %s", charset)
		}
	}

	err = decoder.Decode(&diadoc)
	if err != nil {
		return nil, err
	}

	return diadoc, nil
}

// parseLocalZipFiles читает и распаковывает найденные zip файлы из локальной директории
func parseLocalZipFiles() error {
	return filepath.Walk(filepath.Join(config.GetString("source.local.rootPath"), "zip"), func(path string, info fs.FileInfo, err error) error {
		if err == nil && filepath.Ext(path) == ".zip" {
			slog.Info(fmt.Sprintf("Найден архив %s", info.Name()))

			files, err := zip.OpenReader(path)
			if err == nil {
				slog.Debug(fmt.Sprintf("Обработка архива %s", info.Name()))

				errCount := manageFilesInZip(files)

				files.Close()
				if errCount > 0 {
					slog.Warn("В архиве есть ошибочные документы. Перенос архива в дриеткорию error")
					if err := os.Rename(path, filepath.Join(config.GetString("source.local.rootPath"), "error", info.Name())); err != nil {
						slog.Error("Ошибка при переносе архива в error", slog.String("error", err.Error()))
					}
				} else {
					slog.Info(fmt.Sprintf("Архив %s успешно обработан", info.Name()))
					if err := os.Rename(path, filepath.Join(config.GetString("source.local.rootPath"), "done", info.Name())); err != nil {
						slog.Error("Ошибка при переносе архива в done", slog.String("error", err.Error()))
					}
				}
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

// prepare подгатавливает приложение к работе с источниками
func prepareSource() error {
	var err error
	slog.Debug("Определение типа источнка")
	switch config.GetString("source.type") {
	case "local":
		err = prepareLocalDirs()
	default:
		return fmt.Errorf("не поддерживаемый тип источника %s", config.GetString("source.type"))
	}
	slog.Debug("Подключение к базе данных")
	return err
}
