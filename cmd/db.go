package main

import (
	"context"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"log"
	"time"
)

func initDB(ctx context.Context, url string) (*sqlx.DB, error) {
	conf, errParse := sqlx.Connect("postgres", url)
	if errParse != nil {
		return nil, errParse
	}
	conf.DB.SetMaxIdleConns(20)
	conf.DB.SetConnMaxLifetime(10 * time.Minute)
	if errPing := conf.Ping(); errPing != nil {
		return nil, errPing
	}
	log.Print("pong")
	return conf, nil
}
