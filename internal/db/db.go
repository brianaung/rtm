package db

import (
	"context"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Database struct {
	db *pgxpool.Pool
}

func Init() (*Database, error) {
	dbpool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	err = dbpool.Ping(context.Background())
	if err != nil {
		return nil, err
	}
	return &Database{db: dbpool}, nil
}

func (d *Database) Get() *pgxpool.Pool {
	return d.db
}

func (d *Database) Close() {
	d.db.Close()
}
