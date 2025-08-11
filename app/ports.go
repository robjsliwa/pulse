package app

import (
    "context"
    "github.com/robjsliwa/pulse/domain"
)

// PollRepository defines persistence operations for polls and related aggregates.
type PollRepository interface {
    Create(ctx context.Context, p *domain.Poll) error
    Update(ctx context.Context, p *domain.Poll) error
    Delete(ctx context.Context, id uint) error
    GetByID(ctx context.Context, id uint) (*domain.Poll, error)
    List(ctx context.Context, offset, limit int) ([]domain.Poll, error)

    AddOption(ctx context.Context, opt *domain.Option) error
    ListOptions(ctx context.Context, pollID uint) ([]domain.Option, error)

    CreateVote(ctx context.Context, v *domain.Vote) error
    CountVotesByOption(ctx context.Context, pollID uint) (map[uint]int, int, error)
}

// ResultsStreamer pushes results updates for a poll.
type ResultsStreamer interface {
    Broadcast(pollID uint, res domain.Results)
    Subscribe(pollID uint) (<-chan domain.Results, func())
}

// WebhookDispatcher dispatches signed webhook events.
type WebhookDispatcher interface {
    Dispatch(ctx context.Context, event string, payload any) error
}

