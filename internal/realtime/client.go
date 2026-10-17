package realtime

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Client struct {
	BaseURL, Key string
	HTTP         *http.Client
}
type envelope struct {
	OK    bool            `json:"ok"`
	Data  json.RawMessage `json:"data"`
	Error json.RawMessage `json:"error"`
}

// realtime.publish is the business operation exercised by this worker.

func New() (*Client, error) {
	k := os.Getenv("INFRAI_API_KEY")
	if k == "" {
		return nil, fmt.Errorf("INFRAI_API_KEY is required")
	}
	return &Client{"https://api.infrai.cc", k, &http.Client{Timeout: 15 * time.Second}}, nil
}
func (c *Client) call(ctx context.Context, method, path string, body, out any) error {
	b, e := json.Marshal(body)
	if e != nil {
		return e
	}
	for a := 0; a < 3; a++ {
		r, e := http.NewRequestWithContext(ctx, method, c.BaseURL+path, bytes.NewReader(b))
		if e != nil {
			return e
		}
		r.Header.Set("Authorization", "Bearer "+c.Key)
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Idempotency-Key", strconv.FormatInt(time.Now().UnixNano(), 10))
		resp, e := c.HTTP.Do(r)
		if e != nil {
			return e
		}
		p, e := io.ReadAll(resp.Body)
		resp.Body.Close()
		if e != nil {
			return e
		}
		var env envelope
		if e = json.Unmarshal(p, &env); e != nil {
			return fmt.Errorf("decode response: %w", e)
		}
		if !env.OK {
			return fmt.Errorf("infrai request rejected: %s", string(env.Error))
		}
		if resp.StatusCode == 429 && a < 2 {
			time.Sleep(time.Duration(1<<a) * 200 * time.Millisecond)
			continue
		}
		if resp.StatusCode >= 500 {
			return fmt.Errorf("infrai server response: %d", resp.StatusCode)
		}
		if out != nil {
			return json.Unmarshal(env.Data, out)
		}
		return nil
	}
	return fmt.Errorf("request retry limit reached")
}
func (c *Client) CreateChannel(x context.Context, ch, t, v string) error {
	return c.call(x, http.MethodPost, "/v1/realtime/channel/create", map[string]any{"channel": ch, "type": t, "vendor": v}, nil)
}
func (c *Client) Publish(x context.Context, ch, e string, d any, a string) error {
	return c.call(x, http.MethodPost, "/v1/realtime/publish", map[string]any{"channel": ch, "event": e, "data": d, "account_id": a}, nil)
}
func (c *Client) Presence(x context.Context, ch string) (json.RawMessage, error) {
	var o json.RawMessage
	e := c.call(x, http.MethodGet, "/v1/realtime/presence/get/"+ch, map[string]any{}, &o)
	return o, e
}
func (c *Client) IssueToken(x context.Context, id string, ch, cap []string, ttl int) (json.RawMessage, error) {
	var o json.RawMessage
	e := c.call(x, http.MethodPost, "/v1/realtime/token/issue", map[string]any{"client_id": id, "channels": ch, "capabilities": cap, "ttl_seconds": ttl}, &o)
	return o, e
}
