package main

import (
	"chat-room-media/internal/realtime"
	"context"
	"fmt"
	"log"
	"os"
	"time"
)

type message struct {
	Text    string `json:"text"`
	AssetID string `json:"asset_id"`
}

func main() {
	c, e := realtime.New()
	if e != nil {
		log.Fatal(e)
	}
	room := os.Getenv("CHAT_ROOM")
	if room == "" {
		room = "media-room"
	}
	acct := os.Getenv("ACCOUNT_ID")
	if acct == "" {
		acct = "creator-pipeline"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if e = c.CreateChannel(ctx, room, "presence", "pusher"); e != nil {
		log.Fatal(e)
	}
	if e = c.Publish(ctx, room, "asset.ready", message{"asset processed", "asset-001"}, acct); e != nil {
		log.Fatal(e)
	}
	p, e := c.Presence(ctx, room)
	if e != nil {
		log.Fatal(e)
	}
	fmt.Printf("published asset.ready to %s; presence=%s\n", room, string(p))
}
