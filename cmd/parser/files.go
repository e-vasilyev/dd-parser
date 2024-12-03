package main

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"log/slog"

	"golang.org/x/text/encoding/charmap"
)

// prepareSource подгатавливает приложение к работе с источниками
func prepareSource() error {
	var err error

	slog.Debug("Определение типа источнка")

	switch config.GetString("source.type") {
	case "local":
		err = prepareLocalDirs()
	case "s3":
		err = prepareS3()
	default:
		return fmt.Errorf("не поддерживаемый тип источника %s", config.GetString("source.type"))
	}

	return err
}

// parseZipFiles обрабатывает найденные zip файлы
func parseZipFiles() error {
	var err error

	slog.Debug("Проверка наличия файлов")

	switch config.GetString("source.type") {
	case "local":
		err = parseLocalZipFiles()
	case "s3":
		err = parseS3ZipFiles()
	default:
		return fmt.Errorf("не поддерживаемый тип источника %s", config.GetString("source.type"))
	}

	return err
}

// manageFilesInZip обрабатывает файлы в zip архиве.
// Возвращает количество ошибочных документов.
func manageFilesInZip(zFiles []*zip.File) int {
	var countErr int

	slog.Info(fmt.Sprintf("Найдено %v документов в архиве", len(zFiles)))

	for _, file := range zFiles {
		document, err := parseXML(file)
		if err != nil {
			slog.Error(fmt.Sprintf("Ошибка разбора XML: %s", err.Error()))
			countErr++
		} else {
			slog.Debug(fmt.Sprintf("Разбор документа %s успешно завершен", file.Name))
			if err := pdb.insertDocument(document); err != nil {
				countErr++
			}
		}
	}

	slog.Info(fmt.Sprintf("Обработано %v из %v документов в архиве", len(zFiles)-countErr, len(zFiles)))

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
