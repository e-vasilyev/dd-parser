# DD-PARSER

Приложение dd-parser извлекает данные из xml файлов Контур.Диадок. Полученные данные записываются в базу данных postgresql.

## Сборка приложения

```
go build -v -o dd-parser ./cmd/parser
```

## Настройка

Настройка приложения осуществляется при помощи переменных среды.

| Переменная                 | Значение по умолчанию | Описание                                                                           |
| -------------------------- | --------------------- | ---------------------------------------------------------------------------------- |
| DDP_DATABASE_NAME          | dd_parser             | Имя базы данных                                                                    |
| DDP_DATABASE_HOST          | localhost             | Хост базы данных                                                                   |
| DDP_DATABASE_PORT          | 5432                  | Порт базы данных                                                                   |
| DDP_DATABASE_USERNAME      | postgres              | Пользователь базы данных                                                           |
| DDP_DATABASE_PASSWORD      | postgres              | Пароль пользователя базы данных                                                    |
| DDP_SOURCE_TYPE            | local                 | Тип источника. Поддерживаемые источники: local, s3                                 |
| DDP_SOURCE_LOCAL_ROOT_PATH | source                | Путь до корневой директории локального источника                                   |
| DDP_SOURCE_S3_BUCKET_NAME  | diadoc                | Имя бакета в s3                                                                    |
| DDP_SOURCE_S3_ENDPOINT     | localhost:9000        | Хост и порт до s3                                                                  |
| DDP_SOURCE_S3_USER         | root                  | Пользователь для доступа к s3 (AccessKeyID)                                        |
| DDP_SOURCE_S3_PASSWORD     | password              | Пароль пользователя s3 (SecretAccessKey)                                           |
| DDP_SOURCE_S3_USE_SSL      | false                 | Использовать доступ по SSL                                                         |
| DDP_SOURCE_S3_USE_ROOT     | false                 | Используется учетная записть root. Если true, то bucket будет создан автоматически |
| DDP_LOG_LEVEL              | Info                  | Уровень логирования. Доступные уровни: Info, Warn, Error, Debug                    |

## Использование

Приложение работает с zip архивами полученными на сайте diadoc.kontur.ru. 

1. Для получения архива выберите нужные документы на сайте и нажмите Скачать - Документ в исходном формате.
2. Поместите полученный архив в директорию zip (source/zip для local, zip для s3).
3. После успешной обработки файл будет перемещен в директорию done. 