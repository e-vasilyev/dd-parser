package main

import (
	"archive/zip"
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// S3 описывает подключение к S3 (minio)
type S3 struct {
	client     *minio.Client
	bucketName string
}

// prepareS3 подготавливает бакет и директории для работы
func prepareS3() error {
	var paths = [3]string{"zip", "error", "done"}

	slog.Info("Выбран источник s3")

	slog.Debug(fmt.Sprintf("Подключение к %s", config.GetString("source.s3.endpoint")))

	client, err := minio.New(config.GetString("source.s3.endpoint"), &minio.Options{
		Creds:  credentials.NewStaticV4(config.GetString("source.s3.user"), config.GetString("source.s3.password"), ""),
		Secure: config.GetBool("source.s3.use_ssl"),
	})

	if err != nil {
		return err
	}

	s3.client = client
	s3.bucketName = config.GetString("source.s3.bucket_name")

	if ok := config.GetBool("source.s3.use_root"); ok {
		slog.Debug(fmt.Sprintf("Создание бакета %s", config.GetString("source.s3.baketName")))

		if err := s3.createBucket(s3.bucketName); err != nil {
			return err
		}
	}

	for _, path := range paths {
		object := strings.NewReader("")

		_, err := s3.client.PutObject(
			context.Background(), s3.bucketName, fmt.Sprintf("%s/do_not_delete", path), object, object.Size(), minio.PutObjectOptions{},
		)
		if err != nil {
			return err
		}
	}

	return nil
}

// parseS3ZipFiles читает и распаковывает найденные zip файлы из s3
func parseS3ZipFiles() error {
	opts := minio.ListObjectsOptions{
		Prefix: "zip/",
	}

	for object := range s3.client.ListObjects(context.Background(), s3.bucketName, opts) {
		if strings.HasSuffix(object.Key, ".zip") {
			slog.Info(fmt.Sprintf("Найден архив %s", object.Key))
			if readObject, err := s3.client.GetObject(context.Background(), s3.bucketName, object.Key, minio.GetObjectOptions{}); err == nil {
				if readObjectInfo, err := readObject.Stat(); err == nil {
					if reader, err := zip.NewReader(readObject, readObjectInfo.Size); err == nil {
						slog.Debug(fmt.Sprintf("Обработка архива %s", readObjectInfo.Key))

						errCount := manageFilesInZip(reader.File)

						if errCount > 0 {
							slog.Warn("В архиве есть ошибочные документы. Перенос архива в дриеткорию error")

							if err := s3.moveObject(readObjectInfo.Key, strings.Replace(readObjectInfo.Key, "zip", "error", 1)); err != nil {
								slog.Error(fmt.Sprintf("Ошибка при переносе архива в error: %s", err.Error()))
							}
						} else {
							slog.Info(fmt.Sprintf("Архив %s успешно обработан", readObjectInfo.Key))

							if err := s3.moveObject(readObjectInfo.Key, strings.Replace(readObjectInfo.Key, "zip", "done", 1)); err != nil {
								slog.Error(fmt.Sprintf("Ошибка при переносе архива в done: %s", err.Error()))
							}
						}
					} else {
						slog.Error(fmt.Sprintf("Ошибка при чтении архива: %s", err.Error()))
					}
				} else {
					slog.Error(fmt.Sprintf("Ошибка при получении информации по архиву: %s", err.Error()))
				}
				readObject.Close()
			} else {
				slog.Error(fmt.Sprintf("Ошибка при получении архива: %s", err.Error()))
			}
		}
	}

	return nil
}

// createBucket создает бакет
func (s *S3) createBucket(name string) error {
	if exists, err := s.client.BucketExists(context.Background(), name); err != nil || exists {
		return err
	}

	if err := s.client.MakeBucket(context.Background(), name, minio.MakeBucketOptions{}); err != nil {
		return err
	}

	slog.Info(fmt.Sprintf("Создан бакет %s", name))

	return nil
}

// moveObject переносит объект
func (s *S3) moveObject(src, dest string) error {
	if _, err := s.client.CopyObject(
		context.Background(),
		minio.CopyDestOptions{Bucket: s.bucketName, Object: dest},
		minio.CopySrcOptions{Bucket: s.bucketName, Object: src},
	); err != nil {
		return err
	}

	if err := s.client.RemoveObject(
		context.Background(), s3.bucketName, src, minio.RemoveObjectOptions{},
	); err != nil {
		return err
	}

	return nil
}
