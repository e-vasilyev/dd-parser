package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/e-vasilyev/dd-parser/assets"
	"github.com/pressly/goose/v3"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

// PDB описывает подключение к базе данных postgresql
type PDB struct {
	pool *pgxpool.Pool
}

// connectToDB открывает пул соединений
func connectToDB() (*pgxpool.Pool, error) {
	var url = config.GetString("database.url")
	pool, err := pgxpool.New(context.Background(), url)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(context.Background()); err != nil {
		return nil, err
	}
	return pool, nil
}

// migrationDB запускает миграции баз данных
func (pdb *PDB) migration() error {
	var migrations = &assets.Migrations

	slog.Info("Миграция базы данных")

	goose.SetBaseFS(migrations)
	goose.SetTableName("diadoc_goose_db_version")
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	db := stdlib.OpenDBFromPool(pdb.pool)
	if err := goose.Up(db, "migrations"); err != nil {
		return err
	}

	return nil
}

// insertDocument добавляет записи в БД по документу
func (pdb *PDB) insertDocument(d *DiaDocXML) error {
	var complited bool

	err := pdb.pool.QueryRow(context.Background(), "SELECT complited FROM diadoc_files WHERE id=$1", d.FileId).Scan(&complited)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		_, err := pdb.pool.Exec(
			context.Background(),
			"INSERT INTO diadoc_files (id, form_version, prog_version, complited, document_date, seller, timestamp) VALUES ($1, $2, $3, $4, $5, $6, $7)",
			d.FileId, d.FormVer, d.ProgVer, false, d.Document.Date.Format("2006-01-02"), d.Document.Seller, time.Now().UTC().Format("2006-01-02 03:04:05"),
		)
		if err != nil {
			slog.Error("Ошибка при записи документа в базу данных", slog.String("error", err.Error()))
			return err
		} else {
			slog.Debug("Создана запись в БД для документа")
			return pdb.insertDocumentTable(d)
		}
	case err == nil && !complited:
		return pdb.insertDocumentTable(d)
	case err == nil && complited:
		slog.Warn(fmt.Sprintf("Документ %s обработан ранее. Будет проущен", d.FileId))
	default:
		slog.Error("Ошибка получения id документа", slog.String("error", err.Error()))
		return err
	}
	return nil
}

func (pdb *PDB) insertDocumentTable(d *DiaDocXML) error {
	slog.Info(fmt.Sprintf("Добавление данных товара из документа %s в базу данных", d.FileId))

	tx, err := pdb.pool.Begin(context.Background())
	if err != nil {
		slog.Error("Ошибка при создании транзакции", slog.String("error", err.Error()))
		return err
	}

	defer tx.Rollback(context.Background())

	for _, p := range d.Document.Products {
		price := p.TotalPrice / float32(p.Count)

		_, err := tx.Exec(
			context.Background(),
			"INSERT INTO diadoc_products (id, name, price, file_id, timestamp) VALUES ($1, $2, $3, $4, $5)",
			p.ExtInfo.Code, p.Name, price, d.FileId, time.Now().UTC().Format("2006-01-02 03:04:05"),
		)
		if err != nil {
			slog.Error("Ошибка записи товара в базу данных", slog.String("error", err.Error()))
			return err
		}
	}
	_, err = tx.Exec(context.Background(), "UPDATE diadoc_files SET complited=$2 WHERE id=$1", d.FileId, true)
	if err != nil {
		slog.Error("Ошибка при обновлении документа в базе данных", slog.String("error", err.Error()))
		return err
	} else {
		if err := tx.Commit(context.Background()); err != nil {
			slog.Error("Ошибка при коммите изменений", slog.String("error", err.Error()))
			return err
		} else {
			slog.Info(fmt.Sprintf("Документ %s успешно добавлен в БД", d.FileId))
		}
	}
	return nil
}
