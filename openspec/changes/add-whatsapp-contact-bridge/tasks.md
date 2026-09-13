## 1. Setup and configuration

- [x] 1.1 Add `modernc.org/sqlite` to go.mod and verify the build
- [x] 1.2 Extend internal/config/config.go with WHATSAPP_TOKEN, WHATSAPP_PHONE_NUMBER_ID, WHATSAPP_APP_SECRET, WHATSAPP_VERIFY_TOKEN, WHATSAPP_TEMPLATE_NAME, OWNER_PHONE_NUMBER, DB_PATH (default data.db)
- [x] 1.3 Document all new env vars in .env.example

## 2. SQLite store

- [x] 2.1 Create internal/store with Open(path): open DB, enable WAL, create tables (contacts, replies, webhook_events) via CREATE TABLE IF NOT EXISTS
- [x] 2.2 Implement CreateContact (returns id) and UpdateContactWhatsAppResult(id, notifyMsgID, status)
- [x] 2.3 Implement contact lookups: GetContactByNotifyMsgID, GetContactByID, LatestContact
- [x] 2.4 Implement InsertWebhookEvent (idempotent by wamid, reports whether new) and InsertReply

## 3. WhatsApp service

- [x] 3.1 Create internal/services/whatsapp.go with Cloud API client: SendText(to, body) and SendTemplate(to, name, params) returning the message ID, using POST /v{phone_number_id}/messages
- [x] 3.2 Implement NotifyOwner(contact) — send template when WHATSAPP_TEMPLATE_NAME is set, free-form text otherwise; body includes "#<id>", name, message; skip+log when credentials/owner number are unset; surface error codes on failure
- [x] 3.3 Define webhook payload structs (messages array with from, id, text.body, context.message_id; statuses array ignored)
- [x] 3.4 Add SendReplyEmail to internal/services/email.go for forwarding owner replies to a contact's email

## 4. Webhook handler

- [x] 4.1 Create internal/handlers/webhook.go with GET verification (hub.mode/hub.verify_token/hub.challenge; 403 otherwise)
- [x] 4.2 Implement POST handler: read raw body, verify X-Hub-Signature-256 (constant-time HMAC-SHA256; skip with warning when app secret unset; 401 on mismatch)
- [x] 4.3 Implement event dedupe: insert into webhook_events by wamid, ignore redeliveries
- [x] 4.4 Implement owner-reply correlation cascade: context.message_id → notify_msg_id, "#<id>" body prefix, fallback LatestContact; record match_method; log-and-ignore non-owner senders and empty-contacts case
- [x] 4.5 Forward correlated reply via SendReplyEmail and record in replies; always return 200 after processing

## 5. Contact handler and server wiring

- [x] 5.1 Modify internal/handlers/contact.go: save to store first, then best-effort existing email forward and WhatsApp NotifyOwner; update contact row with notify_msg_id/wa_status; return 200 once saved
- [x] 5.2 Update cmd/server/main.go: construct store and WhatsApp service, inject into handlers, register GET/POST /webhook/whatsapp
- [x] 5.3 Exempt /webhook/* from the global rate limiter while keeping it on /api/* routes

## 6. Verification

- [x] 6.1 Build passes (go build ./...) and server starts with fresh and existing DB files
- [x] 6.2 Manual test: GET /webhook/whatsapp verification with correct and wrong tokens
- [x] 6.3 Manual test: signed POST /webhook/whatsapp (valid HMAC via curl, tampered signature rejected 401, duplicate delivery ignored)
- [x] 6.4 Manual test: POST /api/contact saves a row, sends forward email, and sends WhatsApp to owner (or logs skip/failure)
- [x] 6.5 Manual test: owner reply via quoted context.message_id, "#<id>" prefix, and untagged fallback each forward to the correct contact email exactly once
