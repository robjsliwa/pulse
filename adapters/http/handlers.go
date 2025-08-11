package httpadp

import (
    "net/http"
    "strconv"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/robjsliwa/pulse/app"
    "github.com/robjsliwa/pulse/domain"
)

type Handler struct {
    svc     *app.Service
    stream  app.ResultsStreamer
}

func NewHandler(svc *app.Service, stream app.ResultsStreamer) *Handler { return &Handler{svc: svc, stream: stream} }

// CreatePoll godoc
// @Summary Create a poll
// @Tags polls
// @Accept json
// @Produce json
// @Param payload body CreatePollRequest true "Poll"
// @Success 201 {object} domain.Poll
// @Failure 400 {object} gin.H
// @Router /polls [post]
func (h *Handler) CreatePoll(c *gin.Context) {
    var req CreatePollRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    p := domain.Poll{Title: req.Title, Description: req.Description, Threshold: req.Threshold}
    for _, o := range req.Options { p.Options = append(p.Options, domain.Option{Text: o.Text}) }
    res, err := h.svc.CreatePoll(c.Request.Context(), p)
    if err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return }
    c.JSON(http.StatusCreated, res)
}

// GetPoll godoc
// @Summary Get a poll
// @Tags polls
// @Produce json
// @Param id path int true "Poll ID"
// @Success 200 {object} domain.Poll
// @Failure 404 {object} gin.H
// @Router /polls/{id} [get]
func (h *Handler) GetPoll(c *gin.Context) {
    id, _ := strconv.Atoi(c.Param("id"))
    p, err := h.svc.GetPoll(c.Request.Context(), uint(id))
    if err != nil { c.JSON(http.StatusNotFound, gin.H{"error": err.Error()}); return }
    c.JSON(http.StatusOK, p)
}

// ListPolls godoc
// @Summary List polls
// @Tags polls
// @Produce json
// @Param offset query int false "Offset"
// @Param limit query int false "Limit"
// @Success 200 {array} domain.Poll
// @Router /polls [get]
func (h *Handler) ListPolls(c *gin.Context) {
    offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
    res, err := h.svc.ListPolls(c.Request.Context(), offset, limit)
    if err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return }
    c.JSON(http.StatusOK, res)
}

// UpdatePoll godoc
// @Summary Update a poll
// @Tags polls
// @Accept json
// @Produce json
// @Param id path int true "Poll ID"
// @Param payload body UpdatePollRequest true "Poll"
// @Success 200 {object} domain.Poll
// @Failure 400 {object} gin.H
// @Router /polls/{id} [patch]
func (h *Handler) UpdatePoll(c *gin.Context) {
    id, _ := strconv.Atoi(c.Param("id"))
    var req UpdatePollRequest
    if err := c.ShouldBindJSON(&req); err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return }
    p := domain.Poll{ID: uint(id)}
    if req.Title != nil { p.Title = *req.Title }
    if req.Description != nil { p.Description = *req.Description }
    if req.Threshold != nil { p.Threshold = *req.Threshold }
    res, err := h.svc.UpdatePoll(c.Request.Context(), p)
    if err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return }
    c.JSON(http.StatusOK, res)
}

// DeletePoll godoc
// @Summary Delete a poll
// @Tags polls
// @Param id path int true "Poll ID"
// @Success 204
// @Router /polls/{id} [delete]
func (h *Handler) DeletePoll(c *gin.Context) {
    id, _ := strconv.Atoi(c.Param("id"))
    if err := h.svc.DeletePoll(c.Request.Context(), uint(id)); err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return }
    c.Status(http.StatusNoContent)
}

// ClosePoll godoc
// @Summary Close a poll
// @Tags polls
// @Param id path int true "Poll ID"
// @Success 200 {object} domain.Poll
// @Router /polls/{id}/close [post]
func (h *Handler) ClosePoll(c *gin.Context) {
    id, _ := strconv.Atoi(c.Param("id"))
    res, err := h.svc.ClosePoll(c.Request.Context(), uint(id))
    if err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return }
    c.JSON(http.StatusOK, res)
}

// AddOption godoc
// @Summary Add an option to poll
// @Tags options
// @Accept json
// @Produce json
// @Param id path int true "Poll ID"
// @Param payload body CreateOption true "Option"
// @Success 201 {object} domain.Option
// @Router /polls/{id}/options [post]
func (h *Handler) AddOption(c *gin.Context) {
    id, _ := strconv.Atoi(c.Param("id"))
    var req CreateOption
    if err := c.ShouldBindJSON(&req); err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return }
    opt, err := h.svc.AddOption(c.Request.Context(), uint(id), req.Text)
    if err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return }
    c.JSON(http.StatusCreated, opt)
}

// ListOptions godoc
// @Summary List options for a poll
// @Tags options
// @Produce json
// @Param id path int true "Poll ID"
// @Success 200 {array} domain.Option
// @Router /polls/{id}/options [get]
func (h *Handler) ListOptions(c *gin.Context) {
    id, _ := strconv.Atoi(c.Param("id"))
    opts, err := h.svc.ListOptions(c.Request.Context(), uint(id))
    if err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return }
    c.JSON(http.StatusOK, opts)
}

// Vote godoc
// @Summary Cast a vote
// @Tags votes
// @Accept json
// @Produce json
// @Param id path int true "Poll ID"
// @Param payload body VoteRequest true "Vote"
// @Success 201 {object} domain.Vote
// @Router /polls/{id}/votes [post]
func (h *Handler) Vote(c *gin.Context) {
    id, _ := strconv.Atoi(c.Param("id"))
    var req VoteRequest
    if err := c.ShouldBindJSON(&req); err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return }
    v, err := h.svc.Vote(c.Request.Context(), uint(id), req.OptionID, req.UserID)
    if err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return }
    c.JSON(http.StatusCreated, v)
}

// Results godoc
// @Summary Current poll results
// @Tags results
// @Produce json
// @Param id path int true "Poll ID"
// @Success 200 {object} domain.Results
// @Router /polls/{id}/results [get]
func (h *Handler) Results(c *gin.Context) {
    id, _ := strconv.Atoi(c.Param("id"))
    res, err := h.svc.Results(c.Request.Context(), uint(id))
    if err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return }
    c.JSON(http.StatusOK, res)
}

// ResultsStream godoc
// @Summary Stream poll results via SSE
// @Tags results
// @Produce text/event-stream
// @Param id path int true "Poll ID"
// @Router /polls/{id}/results/stream [get]
func (h *Handler) ResultsStream(c *gin.Context) {
    c.Writer.Header().Set("Content-Type", "text/event-stream")
    c.Writer.Header().Set("Cache-Control", "no-cache")
    c.Writer.Header().Set("Connection", "keep-alive")
    c.Writer.Header().Set("X-Accel-Buffering", "no")
    id, _ := strconv.Atoi(c.Param("id"))
    ch, cancel := h.stream.Subscribe(uint(id))
    defer cancel()

    // send initial
    if res, err := h.svc.Results(c.Request.Context(), uint(id)); err == nil {
        sseWrite(c, res)
    }

    notify := c.Request.Context().Done()
    heartbeat := make(chan struct{}, 1)
    go func() {
        for {
            select {
            case <-notify:
                return
            case <-heartbeat:
                return
            case res := <-ch:
                sseWrite(c, res)
            case <-timeAfter(15): // heartbeat every 15s
                sseWriteComment(c, ":keepalive")
            }
        }
    }()
    <-notify
}

func sseWrite(c *gin.Context, v any) {
    c.SSEvent("message", v)
    c.Writer.Flush()
}

func sseWriteComment(c *gin.Context, msg string) {
    c.Writer.Write([]byte(msg + "\n\n"))
    c.Writer.Flush()
}

// timeAfter split for testability without globals
var timeAfter = func(seconds int) <-chan time.Time { return time.After(time.Duration(seconds) * time.Second) }
