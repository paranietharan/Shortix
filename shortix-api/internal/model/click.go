package model

import (
	"time"

	"github.com/google/uuid"
)

type Click struct {
	ID        int64     `json:"id" gorm:"primary_key;autoIncrement"`
	URLID     uuid.UUID `json:"url_id" gorm:"type:uuid;not null;index"`
	ClickedAt time.Time `json:"clicked_at" gorm:"index;autoCreateTime"`
	IPAddress string    `json:"ip_address" gorm:"type:varchar(45)"`
	UserAgent string    `json:"user_agent" gorm:"type:text"`
	Device    string    `json:"device" gorm:"type:varchar(50)"`
	Referrer  string    `json:"referrer" gorm:"type:text"`
}
