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

	err := pdb.pool.QueryRow(context.Background(), "SELECT complited FROM diadoc_files WHERE id=$1", d.FileID).Scan(&complited)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		_, err := pdb.pool.Exec(
			context.Background(),
			"INSERT INTO diadoc_files (id, form_version, prog_version, complited, invoce_number, document_date, seller, timestamp) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
			d.FileID, d.FormVer, d.ProgVer, false, d.Document.Invoice.Number, d.Document.Invoice.Date.Format("2006-01-02"), d.Document.Seller, time.Now().UTC().Format("2006-01-02 03:04:05"),
		)
		if err != nil {
			slog.Error(fmt.Sprintf("Ошибка при записи документа в базу данных: %s", err.Error()))
			return err
		}

		slog.Debug("Создана запись в БД для документа")
		return pdb.insertDocumentTable(d)

	case err == nil && !complited:
		return pdb.insertDocumentTable(d)
	case err == nil && complited:
		slog.Warn(fmt.Sprintf("Документ %s обработан ранее. Будет проущен", d.FileID))
	default:
		slog.Error(fmt.Sprintf("Ошибка получения id документа: %s", err.Error()))
		return err
	}
	return nil
}

func (pdb *PDB) insertDocumentTable(d *DiaDocXML) error {
	slog.Info(fmt.Sprintf("Добавление данных товара из документа %s в базу данных", d.FileID))

	tx, err := pdb.pool.Begin(context.Background())
	if err != nil {
		slog.Error(fmt.Sprintf("Ошибка при создании транзакции: %s", err.Error()))
		return err
	}

	defer tx.Rollback(context.Background())

	for _, p := range d.Document.Products {
		price := p.TotalPrice / float32(p.Count)

		_, err := tx.Exec(
			context.Background(),
			"INSERT INTO diadoc_products (id, name, count, price, file_id, timestamp) VALUES ($1, $2, $3, $4, $5, $6)",
			p.ExtInfo.Code, p.Name, p.Count, price, d.FileID, time.Now().UTC().Format("2006-01-02 03:04:05"),
		)
		if err != nil {
			slog.Error(fmt.Sprintf("Ошибка записи товара в базу данных: %s", err.Error()))
			return err
		}
	}
	_, err = tx.Exec(context.Background(), "UPDATE diadoc_files SET complited=$2 WHERE id=$1", d.FileID, true)
	if err != nil {
		slog.Error(fmt.Sprintf("Ошибка при обновлении документа в базе данных: %s", err.Error()))
		return err
	}

	if err := tx.Commit(context.Background()); err != nil {
		slog.Error(fmt.Sprintf("Ошибка при коммите изменений: %s", err.Error()))
		return err
	}

	slog.Info(fmt.Sprintf("Документ %s успешно добавлен в БД", d.FileID))

	return nil
}
