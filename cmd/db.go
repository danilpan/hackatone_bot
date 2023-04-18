package main

import (
	"bot/model"
	"context"
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"log"
	"strings"
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

func CheckNumberDb(db sqlx.DB, phone string) error {
	var count int
	query := `SELECT COUNT(*) FROM users WHERE phone_number=$1`
	err := db.Get(&count, query, phone)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	} else {
		return fmt.Errorf("user with such num not found")
	}
}

func AddTgIdToUser(db sqlx.DB, id int64, phone string) error {
	query := `UPDATE users SET tg_id=$1 WHERE phone_number=$2`
	res, err := db.Exec(query, id, phone)
	if err != nil {
		return err
	}
	a, _ := res.RowsAffected()
	if a == 0 {
		return fmt.Errorf("create parking batch row affected 0 -> %w", err)
	}
	return nil
}

func CheckUserDb(db sqlx.DB, id int64) (int64, error) {
	var userId int64
	query := `SELECT id FROM users WHERE tg_id=$1`
	err := db.Get(&userId, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return userId, fmt.Errorf("unf")
		}
		return userId, fmt.Errorf("dbe")
	}
	return userId, nil
}

func CheckUserState(db sqlx.DB, id int64) (int64, error) {
	var userId int64
	query := `SELECT tg_state FROM users WHERE tg_id=$1`
	err := db.Get(&userId, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return userId, fmt.Errorf("unf")
		}
		return userId, fmt.Errorf("dbe")
	}
	return userId, nil
}

func UpdateUserState(db sqlx.DB, id, state int64) error {
	query := `update users SET tg_state=$1 WHERE tg_id=$2`
	res, err := db.Exec(query, state, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("unf")
		}
		return fmt.Errorf("dbe")
	}
	a, _ := res.RowsAffected()
	if a == 0 {
		return fmt.Errorf("dbe")
	}
	return nil
}

func GetActiveUserGuestAccessList(db sqlx.DB, userID int64) (*[]model.WhiteList, error) {
	data := new([]model.WhiteList)
	query := `SELECT id, plate_number, in_black_list, created_at,
       			expires_at, building_id, user_id, is_guest, is_tg_guest 
					FROM white_list 
						WHERE user_id = $1 and (is_guest = true or is_tg_guest = true)
							and expires_at > now()
								order by id DESC;`
	rows, err := db.Queryx(query, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return data, nil
		}
		return data, err
	}
	defer rows.Close()

	for rows.Next() {
		var row model.WhiteList
		if err = rows.Scan(&row.ID, &row.PlateNumber, &row.InBlackList, &row.CreatedAt, &row.ExpiresAt, &row.BuildingID, &row.UserID, &row.IsGuest, &row.IsTgGuest); err != nil {
			return nil, err
		}
		*data = append(*data, row)
	}

	return data, nil
}

func AddUserGuestAccess(db sqlx.DB, wl *model.WhiteList) error {
	if wl.ExpiresAt == "" {
		wl.ExpiresAt = time.Now().AddDate(0, 0, 0).Add(time.Hour * 1).Format("2006-01-02 15:04:05.000000")
	}
	_, err := db.Exec(
		"INSERT INTO white_list (plate_number, expires_at, building_id, user_id, is_guest, is_tg_guest) VALUES($1, $2, $3, $4, $5, $6)",
		strings.ToUpper(wl.PlateNumber), wl.ExpiresAt, wl.BuildingID, wl.UserID, true, true)
	if err != nil {
		return err
	}

	return nil
}

func GetUserBuildings(db sqlx.DB, id int64) ([]model.Building, error) {
	var buildings []model.Building
	query := `select distinct building_id, b.name
from parking p
         join (select id, name from buildings) b on p.building_id = b.id
where is_active = true
  and (user_id = $1 or phone = (select phone_number from users where id = $1));`
	err := db.Select(&buildings, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return buildings, fmt.Errorf("unf")
		}
		return buildings, fmt.Errorf("dbe")
	}
	return buildings, nil
}

func GetUserGuests(db sqlx.DB, id int64) ([]model.WhiteList, error) {
	var guests []model.WhiteList
	query := `select p.id, plate_number, building_id, expires_at
from white_list p
         join (select id, name from buildings) b on p.building_id = b.id
where is_guest = true AND is_tg_guest=true
  and user_id = $1 and expires_at>now();`
	err := db.Select(&guests, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return guests, fmt.Errorf("unf")
		}
		return guests, fmt.Errorf("dbe")
	}
	return guests, nil
}
