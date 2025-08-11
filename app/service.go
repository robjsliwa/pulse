package app

import (
    "context"
    "errors"
    "fmt"
    "time"

    "github.com/robjsliwa/pulse/domain"
)

type Service struct {
    repo      PollRepository
    stream    ResultsStreamer
    webhooks  WebhookDispatcher
    now       func() time.Time
}

func NewService(repo PollRepository, stream ResultsStreamer, webhooks WebhookDispatcher) *Service {
    return &Service{repo: repo, stream: stream, webhooks: webhooks, now: time.Now}
}

// Polls
func (s *Service) CreatePoll(ctx context.Context, p domain.Poll) (*domain.Poll, error) {
    p.Status = domain.PollOpen
    if p.Title == "" || len(p.Options) == 0 {
        return nil, fmt.Errorf("invalid poll: title and options required")
    }
    // ensure no empty options
    opts := make([]domain.Option, 0, len(p.Options))
    for _, o := range p.Options {
        if o.Text == "" {
            return nil, fmt.Errorf("invalid option: text required")
        }
        opts = append(opts, domain.Option{Text: o.Text})
    }
    p.Options = opts
    if err := s.repo.Create(ctx, &p); err != nil {
        return nil, fmt.Errorf("create poll: %w", err)
    }
    return &p, nil
}

func (s *Service) GetPoll(ctx context.Context, id uint) (*domain.Poll, error) {
    p, err := s.repo.GetByID(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("get poll: %w", err)
    }
    return p, nil
}

func (s *Service) ListPolls(ctx context.Context, offset, limit int) ([]domain.Poll, error) {
    ps, err := s.repo.List(ctx, offset, limit)
    if err != nil {
        return nil, fmt.Errorf("list polls: %w", err)
    }
    return ps, nil
}

func (s *Service) UpdatePoll(ctx context.Context, p domain.Poll) (*domain.Poll, error) {
    existing, err := s.repo.GetByID(ctx, p.ID)
    if err != nil {
        return nil, fmt.Errorf("get poll: %w", err)
    }
    if existing.Status == domain.PollClosed {
        return nil, fmt.Errorf("cannot update closed poll")
    }
    if p.Title != "" {
        existing.Title = p.Title
    }
    if p.Description != "" {
        existing.Description = p.Description
    }
    if p.Threshold != 0 {
        existing.Threshold = p.Threshold
    }
    if err := s.repo.Update(ctx, existing); err != nil {
        return nil, fmt.Errorf("update poll: %w", err)
    }
    return existing, nil
}

func (s *Service) DeletePoll(ctx context.Context, id uint) error {
    if err := s.repo.Delete(ctx, id); err != nil {
        return fmt.Errorf("delete poll: %w", err)
    }
    return nil
}

func (s *Service) ClosePoll(ctx context.Context, id uint) (*domain.Poll, error) {
    p, err := s.repo.GetByID(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("get poll: %w", err)
    }
    p.Status = domain.PollClosed
    if err := s.repo.Update(ctx, p); err != nil {
        return nil, fmt.Errorf("close poll: %w", err)
    }
    // webhook event
    _ = s.webhooks.Dispatch(ctx, "poll.closed", map[string]any{"poll_id": p.ID, "title": p.Title})
    return p, nil
}

// Options
func (s *Service) AddOption(ctx context.Context, pollID uint, text string) (*domain.Option, error) {
    p, err := s.repo.GetByID(ctx, pollID)
    if err != nil {
        return nil, fmt.Errorf("get poll: %w", err)
    }
    if p.Status == domain.PollClosed {
        return nil, errors.New("poll is closed")
    }
    if text == "" {
        return nil, fmt.Errorf("option text required")
    }
    opt := &domain.Option{PollID: pollID, Text: text}
    if err := s.repo.AddOption(ctx, opt); err != nil {
        return nil, fmt.Errorf("add option: %w", err)
    }
    return opt, nil
}

func (s *Service) ListOptions(ctx context.Context, pollID uint) ([]domain.Option, error) {
    opts, err := s.repo.ListOptions(ctx, pollID)
    if err != nil {
        return nil, fmt.Errorf("list options: %w", err)
    }
    return opts, nil
}

// Votes and results
func (s *Service) Vote(ctx context.Context, pollID, optionID uint, userID string) (*domain.Vote, error) {
    p, err := s.repo.GetByID(ctx, pollID)
    if err != nil {
        return nil, fmt.Errorf("get poll: %w", err)
    }
    if p.Status == domain.PollClosed {
        return nil, errors.New("poll is closed")
    }
    v := &domain.Vote{PollID: pollID, OptionID: optionID, UserID: userID, CreatedAt: s.now()}
    if err := s.repo.CreateVote(ctx, v); err != nil {
        return nil, fmt.Errorf("create vote: %w", err)
    }
    // recalc results safely from DB (never trust client totals)
    counts, total, err := s.repo.CountVotesByOption(ctx, pollID)
    if err != nil {
        return nil, fmt.Errorf("count votes: %w", err)
    }
    res := domain.Results{PollID: pollID, OptionVotes: counts, Total: total}
    s.stream.Broadcast(pollID, res)

    // webhook vote.created
    _ = s.webhooks.Dispatch(ctx, "vote.created", map[string]any{"poll_id": pollID, "option_id": optionID})

    // threshold check
    if p.Threshold > 0 && total >= p.Threshold {
        _ = s.webhooks.Dispatch(ctx, "poll.threshold_reached", map[string]any{"poll_id": pollID, "threshold": p.Threshold, "total": total})
    }
    return v, nil
}

func (s *Service) Results(ctx context.Context, pollID uint) (domain.Results, error) {
    counts, total, err := s.repo.CountVotesByOption(ctx, pollID)
    if err != nil {
        return domain.Results{}, fmt.Errorf("count votes: %w", err)
    }
    return domain.Results{PollID: pollID, OptionVotes: counts, Total: total}, nil
}

