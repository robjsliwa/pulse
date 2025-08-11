package data

import (
    "fmt"
    "os"

    "github.com/robjsliwa/pulse/adapters/persistence"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
)

// Open opens the SQLite database and runs automigrations.
func Open(dbPath string) (*gorm.DB, error) {
    if dbPath == "" { dbPath = "./pulse.db" }
    // ensure directory exists
    if err := os.MkdirAll(dirname(dbPath), 0o755); err != nil { return nil, fmt.Errorf("mkdir db dir: %w", err) }
    db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
    if err != nil { return nil, fmt.Errorf("open db: %w", err) }
    if err := db.AutoMigrate(&persistence.PollModel{}, &persistence.OptionModel{}, &persistence.VoteModel{}); err != nil {
        return nil, fmt.Errorf("automigrate: %w", err)
    }
    return db, nil
}

func dirname(path string) string {
    i := len(path) - 1
    for i >= 0 && path[i] != '/' { i-- }
    if i <= 0 { return "." }
    return path[:i]
}

