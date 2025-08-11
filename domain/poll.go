package domain

import "time"

type PollStatus string

const (
    PollOpen   PollStatus = "open"
    PollClosed PollStatus = "closed"
)

type Poll struct {
    ID          uint
    Title       string
    Description string
    Status      PollStatus
    Threshold   int // optional threshold to trigger webhook
    Options     []Option
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

type Option struct {
    ID        uint
    PollID    uint
    Text      string
    CreatedAt time.Time
    UpdatedAt time.Time
}

type Vote struct {
    ID        uint
    PollID    uint
    OptionID  uint
    UserID    string // optional identifier
    CreatedAt time.Time
}

// Results represents counts per option.
type Results struct {
    PollID      uint
    OptionVotes map[uint]int
    Total       int
}

