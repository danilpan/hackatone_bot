package model

type WhiteList struct {
	ID          int    `json:"id" db:"id"`
	PlateNumber string `json:"plate_number" db:"plate_number"`
	InBlackList bool   `json:"in_black_list" db:"in_black_list"`
	ChannelID   string `json:"channel_id" db:"channel_id"`
	BuildingID  int    `json:"building_id" db:"building_id"`
	UserID      int    `json:"user_id,omitempty" db:"user_id"`
	CreatedAt   string `json:"created_at" db:"created_at"`
	ExpiresAt   string `json:"expires_at,omitempty" db:"expires_at"`
	IsGuest     bool   `json:"is_guest,omitempty" db:"is_guest"`
	IsTgGuest   bool   `json:"is_tg_guest" db:"is_tg_guest"`
	TotalRows   int    `json:"-" db:"total_rows"`
}

type Building struct {
	Id   int    `json:"id" db:"building_id"`
	Name string `json:"name" db:"name"`
}
