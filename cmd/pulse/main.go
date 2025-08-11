package main

import (
    "log"
    "net/http"
    "os"
    "strings"

    "github.com/gin-contrib/cors"
    "github.com/gin-contrib/requestid"
    "github.com/gin-gonic/gin"
    swaggerFiles "github.com/swaggo/files"
    ginSwagger "github.com/swaggo/gin-swagger"

    httpadp "github.com/robjsliwa/pulse/adapters/http"
    "github.com/robjsliwa/pulse/adapters/persistence"
    "github.com/robjsliwa/pulse/app"
    "github.com/robjsliwa/pulse/data"
    "github.com/robjsliwa/pulse/internal/webhook"
    _ "github.com/robjsliwa/pulse/docs"
)

// @title Pulse API
// @version 0.1.0
// @description Live polls & reactions service.
// @BasePath /

func main() {
    // Env
    port := getenv("PORT", "8080")
    dbPath := getenv("DB_PATH", "./pulse.db")
    corsOrigins := getenv("CORS_ORIGINS", "*")
    maxRetries := atoi(getenv("WEBHOOK_MAX_RETRIES", "5"))
    webhookTargets := splitNonEmpty(getenv("WEBHOOK_TARGETS", ""))
    // Note: webhook secret is optional; if empty, signature is computed with empty key.
    secret := []byte(os.Getenv("WEBHOOK_SECRET"))

    // DB
    db, err := data.Open(dbPath)
    if err != nil { log.Fatalf("db open: %v", err) }

    // Adapters
    repo := persistence.NewRepo(db)
    broadcaster := httpadp.NewBroadcaster()
    dispatcher := webhook.NewDispatcher(webhookTargets, secret, maxRetries)
    svc := app.NewService(repo, broadcaster, dispatcher)

    // HTTP
    r := gin.New()
    r.Use(gin.Recovery())
    r.Use(requestid.New())
    r.Use(corsMiddleware(corsOrigins))
    r.Use(limitBody(1 << 20)) // 1MB payload limit

    h := httpadp.NewHandler(svc, broadcaster)

    // Routes
    r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
    r.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })

    polls := r.Group("/polls")
    {
        polls.POST("", h.CreatePoll)
        polls.GET("", h.ListPolls)
        polls.GET(":id", h.GetPoll)
        polls.PATCH(":id", h.UpdatePoll)
        polls.DELETE(":id", h.DeletePoll)
        polls.POST(":id/close", h.ClosePoll)

        polls.POST(":id/options", h.AddOption)
        polls.GET(":id/options", h.ListOptions)

        polls.POST(":id/votes", h.Vote)
        polls.GET(":id/results", h.Results)
        polls.GET(":id/results/stream", h.ResultsStream)
    }

    log.Printf("Pulse listening on :%s", port)
    if err := r.Run(":" + port); err != nil { log.Fatalf("server error: %v", err) }
}

func getenv(k, def string) string { if v := os.Getenv(k); v != "" { return v }; return def }

func atoi(s string) int { n := 0; for _, ch := range s { if ch < '0' || ch > '9' { continue }; n = n*10 + int(ch-'0') }; return n }

func splitNonEmpty(s string) []string {
    if s == "" { return nil }
    parts := strings.Split(s, ",")
    out := make([]string, 0, len(parts))
    for _, p := range parts { p = strings.TrimSpace(p); if p != "" { out = append(out, p) } }
    return out
}

func corsMiddleware(origins string) gin.HandlerFunc {
    cfg := cors.DefaultConfig()
    if origins == "*" {
        cfg.AllowAllOrigins = true
    } else {
        cfg.AllowOrigins = splitNonEmpty(origins)
    }
    cfg.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"}
    cfg.ExposeHeaders = []string{"Request-Id"}
    cfg.AllowMethods = []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"}
    return cors.New(cfg)
}

func limitBody(maxBytes int64) gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
        c.Next()
    }
}
