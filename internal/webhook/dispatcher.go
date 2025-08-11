package webhook

import (
    "bytes"
    "context"
    "crypto/hmac"
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "time"
)

type Dispatcher struct {
    client        *http.Client
    targets       []string
    secret        []byte
    maxRetries    int
}

func NewDispatcher(targets []string, secret []byte, maxRetries int) *Dispatcher {
    if maxRetries <= 0 { maxRetries = 5 }
    return &Dispatcher{client: &http.Client{Timeout: 10 * time.Second}, targets: targets, secret: secret, maxRetries: maxRetries}
}

func (d *Dispatcher) Dispatch(ctx context.Context, event string, payload any) error {
    if len(d.targets) == 0 {
        return nil // no-op if not configured
    }
    body, err := json.Marshal(payload)
    if err != nil { return fmt.Errorf("marshal payload: %w", err) }
    ts := time.Now().UTC().Format(time.RFC3339)
    sig := d.sign(body)
    for _, url := range d.targets {
        // retry with exponential backoff
        var lastErr error
        for attempt := 0; attempt < d.maxRetries; attempt++ {
            req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
            req.Header.Set("Content-Type", "application/json")
            req.Header.Set("Pulse-Event", event)
            req.Header.Set("Pulse-Timestamp", ts)
            req.Header.Set("Pulse-Signature", sig)
            resp, err := d.client.Do(req)
            if err == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
                _ = resp.Body.Close()
                break
            }
            if resp != nil { _ = resp.Body.Close() }
            lastErr = err
            time.Sleep(time.Duration(1<<attempt) * 200 * time.Millisecond)
        }
        if lastErr != nil {
            log.Printf("webhook dispatch failed for %s: %v", url, lastErr)
        }
    }
    return nil
}

func (d *Dispatcher) sign(body []byte) string {
    mac := hmac.New(sha256.New, d.secret)
    mac.Write(body)
    return hex.EncodeToString(mac.Sum(nil))
}

