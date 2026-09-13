## Why

Contact form submissions currently reach the business owner only by email, which is easy to miss and slow to respond to. The owner lives in WhatsApp, so inquiries should surface there instantly — and replies typed in WhatsApp should flow back to the contact without the owner touching an email client. The backend also has no persistence today, so submissions are lost once the notification email is sent.

## What Changes

- Add a `GET/POST /webhook/whatsapp` endpoint that receives WhatsApp Business Cloud API webhooks (Meta verification handshake, `X-Hub-Signature-256` validation, delivery deduplication).
- Add a SQLite datastore (pure-Go driver, no CGO) persisting contact submissions, owner replies, and raw webhook events.
- Modify `POST /api/contact` to save every submission to SQLite and notify the owner on both channels: the existing forward email (kept) and a new WhatsApp message to the owner's personal number via the Cloud API.
- Add reply forwarding: when the business owner sends any WhatsApp message to the registered business number, the webhook correlates it to the originating contact (quoted-reply message ID → `#<id>` tag → most-recent-contact fallback) and forwards the reply to the contact's email address.
- Add configuration for WhatsApp credentials, owner phone number, and database path.

## Capabilities

### New Capabilities
- `contact-storage`: Persist contact form submissions (name, email, phone, message) in SQLite, with notification message IDs for later reply correlation.
- `whatsapp-webhook`: Receive and validate WhatsApp Cloud API webhook events at `/webhook/whatsapp` — Meta GET verification, HMAC signature verification, idempotent processing, fast 200 responses.
- `owner-notification`: On contact form submission, notify the business owner via WhatsApp (template or free-form text) in addition to the existing email, on a best-effort basis.
- `reply-forwarding`: Correlate any inbound WhatsApp message from the business owner to a stored contact and forward the reply to that contact's email.

### Modified Capabilities

(none — no existing specs)

## Impact

- **New dependency**: `modernc.org/sqlite` (pure Go) added to `go.mod`.
- **New env vars**: `WHATSAPP_TOKEN`, `WHATSAPP_PHONE_NUMBER_ID`, `WHATSAPP_APP_SECRET`, `WHATSAPP_VERIFY_TOKEN`, `WHATSAPP_TEMPLATE_NAME` (optional), `OWNER_PHONE_NUMBER`, `DB_PATH` — documented in `.env.example`.
- **Code**: new `internal/store` (SQLite), new WhatsApp service + webhook handler, modified contact handler and `cmd/server/main.go` wiring.
- **Routes**: new `/webhook/whatsapp` (exempt from the global rate limiter); `/api/contact` behavior extended (still returns 200 once saved).
- **External prerequisites**: Meta app with WhatsApp Business Account (already exists); an approved utility template for reliable owner notifications outside the 24-hour window (to be created during setup); public HTTPS callback URL configured in the Meta App Dashboard.
- **Out of scope**: SMS forwarding (email only for now), messaging the contact back on WhatsApp, any admin UI.
