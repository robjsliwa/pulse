package persistence

import (
    "context"
    "fmt"

    "github.com/robjsliwa/pulse/domain"
    "gorm.io/gorm"
)

type Repo struct {
    db *gorm.DB
}

func NewRepo(db *gorm.DB) *Repo { return &Repo{db: db} }

func (r *Repo) Create(ctx context.Context, p *domain.Poll) error {
    m := PollModel{Title: p.Title, Description: p.Description, Status: string(p.Status), Threshold: p.Threshold}
    for _, o := range p.Options {
        m.Options = append(m.Options, OptionModel{Text: o.Text})
    }
    if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
        return fmt.Errorf("create poll: %w", err)
    }
    p.ID = m.ID
    p.Options = nil // caller should re-fetch if needed
    return nil
}

func (r *Repo) Update(ctx context.Context, p *domain.Poll) error {
    return r.db.WithContext(ctx).Model(&PollModel{ID: p.ID}).Updates(map[string]any{
        "title": p.Title, "description": p.Description, "status": string(p.Status), "threshold": p.Threshold,
    }).Error
}

func (r *Repo) Delete(ctx context.Context, id uint) error {
    return r.db.WithContext(ctx).Delete(&PollModel{}, id).Error
}

func (r *Repo) GetByID(ctx context.Context, id uint) (*domain.Poll, error) {
    var m PollModel
    if err := r.db.WithContext(ctx).Preload("Options").First(&m, id).Error; err != nil {
        return nil, fmt.Errorf("get poll: %w", err)
    }
    p := &domain.Poll{ID: m.ID, Title: m.Title, Description: m.Description, Status: domain.PollStatus(m.Status), Threshold: m.Threshold, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt}
    for _, o := range m.Options {
        p.Options = append(p.Options, domain.Option{ID: o.ID, PollID: o.PollID, Text: o.Text, CreatedAt: o.CreatedAt, UpdatedAt: o.UpdatedAt})
    }
    return p, nil
}

func (r *Repo) List(ctx context.Context, offset, limit int) ([]domain.Poll, error) {
    var ms []PollModel
    q := r.db.WithContext(ctx).Model(&PollModel{}).Order("id DESC").Offset(offset)
    if limit > 0 { q = q.Limit(limit) }
    if err := q.Preload("Options").Find(&ms).Error; err != nil {
        return nil, fmt.Errorf("list polls: %w", err)
    }
    out := make([]domain.Poll, 0, len(ms))
    for _, m := range ms {
        p := domain.Poll{ID: m.ID, Title: m.Title, Description: m.Description, Status: domain.PollStatus(m.Status), Threshold: m.Threshold, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt}
        for _, o := range m.Options { p.Options = append(p.Options, domain.Option{ID: o.ID, PollID: o.PollID, Text: o.Text, CreatedAt: o.CreatedAt, UpdatedAt: o.UpdatedAt}) }
        out = append(out, p)
    }
    return out, nil
}

func (r *Repo) AddOption(ctx context.Context, opt *domain.Option) error {
    m := OptionModel{PollID: opt.PollID, Text: opt.Text}
    if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
        return fmt.Errorf("add option: %w", err)
    }
    opt.ID = m.ID
    return nil
}

func (r *Repo) ListOptions(ctx context.Context, pollID uint) ([]domain.Option, error) {
    var ms []OptionModel
    if err := r.db.WithContext(ctx).Where("poll_id = ?", pollID).Find(&ms).Error; err != nil {
        return nil, fmt.Errorf("list options: %w", err)
    }
    out := make([]domain.Option, 0, len(ms))
    for _, m := range ms { out = append(out, domain.Option{ID: m.ID, PollID: m.PollID, Text: m.Text, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt}) }
    return out, nil
}

func (r *Repo) CreateVote(ctx context.Context, v *domain.Vote) error {
    m := VoteModel{PollID: v.PollID, OptionID: v.OptionID, UserID: v.UserID, CreatedAt: v.CreatedAt}
    if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
        return fmt.Errorf("create vote: %w", err)
    }
    v.ID = m.ID
    return nil
}

func (r *Repo) CountVotesByOption(ctx context.Context, pollID uint) (map[uint]int, int, error) {
    type row struct{ OptionID uint; Cnt int }
    var rows []row
    err := r.db.WithContext(ctx).
        Model(&VoteModel{}).
        Select("option_id, COUNT(*) as cnt").
        Where("poll_id = ?", pollID).
        Group("option_id").
        Scan(&rows).Error
    if err != nil { return nil, 0, fmt.Errorf("count votes: %w", err) }
    res := map[uint]int{}
    total := 0
    for _, rrow := range rows { res[rrow.OptionID] = rrow.Cnt; total += rrow.Cnt }
    return res, total, nil
}

