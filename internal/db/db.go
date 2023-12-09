package db

import (
	"context"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

type database struct {
	db *pgxpool.Pool
}

func Init() (*database, error) {
	dbpool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	err = dbpool.Ping(context.Background())
	if err != nil {
		return nil, err
	}
	return &database{db: dbpool}, nil
}

func (d *database) Get() *pgxpool.Pool {
	return d.db
}

func (d *database) Close() {
	d.db.Close()
}
