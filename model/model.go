package model

import "time"

type Reservation struct {
	Id        int       `json:"id" db:"id"`
	TableId   int       `json:"table_id" db:"table_id"`
	IIN       string    `json:"user_iin" db:"user_iin"`
	TimeFrom  time.Time `json:"time_from" db:"time_from"`
	TimeTo    time.Time `json:"time_to" db:"time_to"`
	Confirmed bool      `json:"confirmed" db:"confirmed"`
	Persons   int       `json:"persons" db:"persons"`
}
