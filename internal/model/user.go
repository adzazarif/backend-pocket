package model

import "time"

type User struct {
    ID        string    `gorm:"type:varchar(36);primaryKey"`
    Name      string    `gorm:"type:varchar(100);not null"`
    Email     string    `gorm:"type:varchar(255);not null;uniqueIndex"`
    Password  string    `gorm:"type:varchar(255);not null"`
    AvatarURL *string   `gorm:"type:varchar(500)"`
    CreatedAt time.Time `gorm:"precision:3;autoCreateTime"`
    UpdatedAt time.Time `gorm:"precision:3;autoUpdateTime"`
}

func (User) TableName() string {
    return "users"
}
