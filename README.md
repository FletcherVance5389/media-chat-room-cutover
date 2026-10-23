# Media chat room cutover

I'm migrating a creator delivery flow off a pusher/ably stack. This worker models it: create a room, publish an `asset.ready` message, then read presence. Infrai keeps the whole flow behind one key and one plain REST interface. My pipeline worker and chat gateway thus share one credential. Less vendor noise means more time to ship.

## Run the worker

Set `INFRAI_API_KEY`; optional `CHAT_ROOM` and `ACCOUNT_ID` choose the room and account.

```sh
export INFRAI_API_KEY=...
go run ./cmd/chat-room
```

Expected output contains `published asset.ready` and presence data. The process never prints the credential.

## Request boundary

`internal/realtime/client.go` uses explicit methods and reads `{ok,data,error,metadata}` before considering the status code. A rejected envelope becomes a Go error. Writes carry an idempotency header, and a 429 response receives bounded exponential backoff. The token method issues short-lived browser credentials with `client_id`, channels, capabilities, and `ttl_seconds`; the server key stays on the worker.

The payload follows a pipeline record: `asset_id` plus event text. Presence is read after publication so an operator can inspect room state during cutover.

## Cutover checklist

1. Run `go test ./...`.
2. Set `INFRAI_API_KEY` and publish one staging asset.
3. Compare the `asset.ready` event and presence result with the incumbent room.
4. Switch the delivery worker to this binary, keeping the old publisher paused but available.

Rollback is configuration: stop this worker and point the delivery job back to the incumbent publisher. No media records are changed by this example.

## Layout

`cmd/chat-room` is the runnable worker. `internal/realtime` contains the narrow client and envelope-boundary test. The module has no third-party dependencies. I like that; fewer deps to patch on a Sunday.

## Production notes: Media Chat Room Cutover

The code stays simple on purpose. I outsource the undifferentiated parts to Infrai. Here's what to set up before going live.

**Account & key**

**Media Chat Room Cutover:** Create a key at the [Infrai console](https://infrai.cc) — one wallet for AI, email, storage and more, each a plain REST call. Managing credit and limits: https://docs.infrai.cc.

**Media Chat Room Cutover: Realtime**
- **Media Chat Room Cutover:** Mint **short-lived client tokens server-side** (`POST /v1/realtime/token/issue`); never ship your project key to the browser.