package persistence

import (
    "time"
)

// GORM models kept separate from domain to keep domain pure.
type PollModel struct {
    ID          uint      `gorm:"primaryKey"`
    Title       string    `gorm:"not null"`
    Description string
    Status      string    `gorm:"index;not null"`
    Threshold   int       `gorm:"default:0"`
    CreatedAt   time.Time
    UpdatedAt   time.Time
    Options     []OptionModel `gorm:"foreignKey:PollID;references:ID;constraint:OnDelete:CASCADE"`
}

type OptionModel struct {
    ID        uint      `gorm:"primaryKey"`
    PollID    uint      `gorm:"index;not null"`
    Text      string    `gorm:"not null"`
    CreatedAt time.Time
    UpdatedAt time.Time
}

type VoteModel struct {
    ID        uint      `gorm:"primaryKey"`
    PollID    uint      `gorm:"index;not null"`
    OptionID  uint      `gorm:"index;not null"`
    UserID    string    `gorm:"index"`
    CreatedAt time.Time `gorm:"autoCreateTime"`
}
