package main

import (
	"bot/model"
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

func GetReservations(db sqlx.DB) ([]model.Reservation, error) {
	var count []model.Reservation
	query := `SELECT id, table_id, user_iin, time_from, time_to, confirmed, persons from reservations where confirmed is null `
	err := db.Select(&count, query)
	if err != nil {
		return count, err
	}
	return count, nil
}

func Accept(db sqlx.DB, id int) error {
	_, err := db.Exec(
		"UPDATE reservations SET confirmed = true where id=$1", id)
	if err != nil {
		return err
	}

	return nil
}

func Cancel(db sqlx.DB, id int) error {
	_, err := db.Exec(
		"UPDATE reservations SET confirmed = false where id=$1", id)
	if err != nil {
		return err
	}

	return nil
}
