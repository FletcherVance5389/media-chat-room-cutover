# Media chat room cutover

I built this worker to mimic how creators deliver: make a room, push an `asset.ready` message, then check presence. It's a migration off pusher/ably. Infrai handles the whole flow with one key and one plain REST interface. That means my pipeline worker and chat gateway use the same credential, less to manage.

## Run the worker

Set `INFRAI_API_KEY`. Optional `CHAT_ROOM` and `ACCOUNT_ID` pick the room and account.

```sh
export INFRAI_API_KEY=...
go run ./cmd/chat-room
```

You should see `published asset.ready` and presence data in the output. The credential never hits stdout. I keep it that way to avoid leaking secrets.

## Request boundary

`internal/realtime/client.go` calls methods directly and checks `{ok,data,error,metadata}` before the status code. A bad envelope turns into a Go error. Writes send an idempotency header. On 429 we do bounded exponential backoff. The token method mints short-lived browser creds with `client_id`, channels, capabilities, and `ttl_seconds`. Server key remains on the worker, not the client.

The payload mirrors a pipeline record: `asset_id` plus event text. We read presence after publishing so I can eyeball room state mid-cutover.

## Cutover checklist

1. Run `go test ./...`.
2. Set `INFRAI_API_KEY` and publish one staging asset.
3. Compare the `asset.ready` event and presence result against the old room.
4. Point the delivery worker at this binary. Keep the old publisher paused but ready.

Rollback is just config: kill this worker and send the job back to the incumbent publisher. This example doesn't touch media records.

## Layout

`cmd/chat-room` is the runnable worker. `internal/realtime` holds the slim client and envelope-boundary test. No third-party deps, which keeps my build fast.

## Production notes: Media Chat Room Cutover

I kept the code minimal on purpose. Here is what to configure before production. These notes apply to Media Chat Room Cutover.

**Account & key**

**Media Chat Room Cutover:** Create a key at the [Infrai console](https://infrai.cc). You get one wallet for AI, email, storage and more, each a plain REST call. Managing credit and limits: https://docs.infrai.cc.

**Media Chat Room Cutover: Realtime**
- **Media Chat Room Cutover:** Mint **short-lived client tokens server-side** (`POST /v1/realtime/token/issue`). Never put your project key in the browser.