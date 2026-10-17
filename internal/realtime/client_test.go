package realtime

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPublishReadsBusinessEnvelope(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/realtime/publish" {
			t.Fatalf("request=%s %s", r.Method, r.URL.Path)
		}
		var b map[string]any
		_ = json.NewDecoder(r.Body).Decode(&b)
		if b["channel"] != "room-7" || b["event"] != "message" {
			t.Fatalf("body=%#v", b)
		}
		w.WriteHeader(400)
		_, _ = w.Write([]byte(`{"ok":false,"error":{"message":"request rejected"}}`))
	}))
	defer s.Close()
	c := &Client{BaseURL: s.URL, Key: "test", HTTP: s.Client()}
	if e := c.Publish(context.Background(), "room-7", "message", map[string]string{"text": "hi"}, "acct"); e == nil {
		t.Fatal("expected envelope error")
	}
}
