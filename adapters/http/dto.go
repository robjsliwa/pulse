package httpadp

// Request/Response DTOs for binding/validation layer.

type CreatePollRequest struct {
    Title       string         `json:"title" binding:"required,min=1,max=200"`
    Description string         `json:"description"`
    Threshold   int            `json:"threshold"`
    Options     []CreateOption `json:"options" binding:"required,dive"`
}

type CreateOption struct {
    Text string `json:"text" binding:"required,min=1,max=200"`
}

type UpdatePollRequest struct {
    Title       *string `json:"title"`
    Description *string `json:"description"`
    Threshold   *int    `json:"threshold"`
}

type VoteRequest struct {
    OptionID uint   `json:"option_id" binding:"required"`
    UserID   string `json:"user_id"`
}

