-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS diadoc_files (
    id varchar(256),
    form_version varchar(64),
    prog_version varchar(64),
    complited boolean NOT NULL,
    invoce_number varchar(64),
    document_date date NOT NULL,
    seller varchar(128) NOT NULL,
    timestamp timestamp NOT NULL,
    PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS diadoc_products (
    id varchar(10),
    name varchar(256) NOT NULL,
    count int NOT NULL,
    price numeric(8, 2) NOT NULL,
    file_id varchar(256),
    timestamp timestamp NOT NULL,
    PRIMARY KEY (id, file_id),
    FOREIGN KEY (file_id) REFERENCES diadoc_files (Id)
);
-- +goose StatementEnd