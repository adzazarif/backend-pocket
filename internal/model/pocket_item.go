package model

import (
    "database/sql/driver"
    "encoding/json"
    "fmt"
    "time"
)

// StringArray adalah custom type untuk JSON column tags
type StringArray []string

func (s StringArray) Value() (driver.Value, error) {
    if s == nil {
        return "[]", nil
    }
    b, err := json.Marshal(s)
    return string(b), err
}

func (s *StringArray) Scan(value interface{}) error {
    if value == nil {
        *s = StringArray{}
        return nil
    }
    bytes, ok := value.([]byte)
    if !ok {
        return fmt.Errorf("cannot scan type %T into StringArray", value)
    }
    return json.Unmarshal(bytes, s)
}

type PocketItem struct {
    ID          string      `gorm:"type:varchar(36);primaryKey"`
    UserID      string      `gorm:"type:varchar(36);not null;index"`
    Title       string      `gorm:"type:varchar(120);not null"`
    URL         *string     `gorm:"type:varchar(2048)"`
    Description *string     `gorm:"type:text"`
    ContentType string      `gorm:"type:enum('article','video','document','note');not null;default:'article'"`
    Status      string      `gorm:"type:enum('unread','reading','read','archived');not null;default:'unread'"`
    IsFavorite  bool        `gorm:"type:tinyint(1);not null;default:0"`
    Tags        StringArray `gorm:"type:json"`
    CreatedAt   time.Time   `gorm:"precision:3;autoCreateTime"`
    UpdatedAt   time.Time   `gorm:"precision:3;autoUpdateTime"`

    User User `gorm:"foreignKey:UserID;references:ID"`
}

func (PocketItem) TableName() string {
    return "pocket_items"
}
