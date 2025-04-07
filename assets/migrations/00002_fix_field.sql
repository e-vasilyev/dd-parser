-- +goose Up
-- +goose StatementBegin
ALTER TABLE diadoc_files RENAME COLUMN invoce_number TO invoice_number;

ALTER TABLE  diadoc_products 
    DROP CONSTRAINT diadoc_products_file_id_fkey,
    ADD CONSTRAINT diadoc_products_file_id_fkey 
    FOREIGN KEY (file_id) REFERENCES diadoc_files(id) ON DELETE CASCADE;
-- +goose StatementEnd