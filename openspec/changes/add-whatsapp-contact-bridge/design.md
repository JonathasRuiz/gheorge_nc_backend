## Context

The backend is a small Gin service (`cmd/server/main.go`) with two routes: `POST /api/contact` (validates name/email/phone/message, forwards an email via `net/smtp` to `CONTACT_FORWARD_TO`) and `POST /api/chat` (OpenAI). There is no database, no WhatsApp integration. A Meta app with a WhatsApp Business Account and registered business number already exists.

Key platform constraint that shapes the whole design: the WhatsApp Cloud API sends messages **as** the registered business number — it cannot send a message *to* that same number. Webhooks only fire when a user **sends a message to** the business number. Therefore the owner reads and replies from their **personal** WhatsApp number, and that thread (owner ↔ business number) acts as a control channel.

```
Visitor ──POST /api/contact──▶ Backend
                                 ├─ 1. save → SQLite (contacts)
                                 ├─ 2. email → CONTACT_FORWARD_TO        (existing)
                                 └─ 3. WhatsApp: business# → OWNER#      (notify, tagged #<id>)
Owner replies in WhatsApp ──▶ Meta ──POST /webhook/whatsapp──▶ Backend
                                 ├─ 4. verify signature + dedupe
                                 ├─ 5. from == OWNER_NUMBER? correlate contact
                                 └─ 6. email reply → contact's email
```

## Goals / Non-Goals

**Goals:**
- Persist every contact form submission in SQLite before any notification is attempted.
- Notify the owner on both email (existing behavior, kept) and WhatsApp (new) for each submission.
- Accept and validate WhatsApp Cloud API webhooks at `/webhook/whatsapp`.
- Treat **any** message from the owner's number as a reply to a contact, with a deterministic correlation cascade.
- Forward the owner's reply text to the matched contact's email address.
- Keep the contact form responsive: 200 once saved; notifications are best-effort.

**Non-Goals:**
- SMS forwarding (email only; SMS provider may be added later).
- Sending WhatsApp messages *to* the contact (visitor).
- Admin UI, message threading UI, or replying from the website.
- Queueing/retry infrastructure for failed notifications (logged only).

## Decisions

**D1 — Topology: owner's personal number as control channel.**
The business number (Cloud API) is the sender for outbound API messages; the owner's personal number receives notifications and sends replies. Alternative considered: owner replying via Meta Business Suite inbox — rejected because business-sent messages do not trigger the received-message webhook this design depends on.

**D2 — Correlation cascade (any owner message is a response).**
In order of reliability:
1. `context.message_id` in the webhook payload matches `contacts.notify_msg_id` (owner used WhatsApp swipe-reply).
2. Message body starts with `#<id>` (owner typed the tag shown in the notification).
3. Fallback: most recently created contact.
Each matched reply records its `match_method` in the `replies` table for audit. Alternative considered: single pending-conversation assumption — rejected as fragile with multiple submissions; prefix-only matching — rejected as requiring owner discipline; the fallback keeps "any message is a response" true in practice.

**D3 — SQLite driver: `modernc.org/sqlite` (pure Go).**
No CGO, keeps cross-compilation and deployment trivial. `mattn/go-sqlite3` rejected due to build friction. Schema (auto-migrated on startup via `CREATE TABLE IF NOT EXISTS`):

```sql
contacts(id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL, email TEXT NOT NULL,
         phone TEXT NOT NULL, message TEXT NOT NULL, notify_msg_id TEXT,
         wa_status TEXT, created_at DATETIME DEFAULT CURRENT_TIMESTAMP)
replies(id INTEGER PRIMARY KEY AUTOINCREMENT, contact_id INTEGER NOT NULL REFERENCES contacts(id),
        wa_message_id TEXT UNIQUE NOT NULL, body TEXT NOT NULL, match_method TEXT NOT NULL,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP)
webhook_events(wa_message_id TEXT PRIMARY KEY, raw TEXT NOT NULL,
               created_at DATETIME DEFAULT CURRENT_TIMESTAMP)
```

**D4 — Owner notification: template if configured, free-form otherwise.**
Free-form API messages to the owner only deliver inside the 24-hour customer-service window; outside it the Cloud API rejects with errors 131047/131026. If `WHATSAPP_TEMPLATE_NAME` is set, send that template (params: `#<id>`, name, message excerpt); otherwise send free-form text and log a template hint on failure. No template is approved yet — creating one (e.g. *"New website enquiry #{{1}} from {{2}}: {{3}} — swipe-reply to respond."*) is a manual setup step, not code.

**D5 — Webhook security and delivery semantics.**
- `GET /webhook/whatsapp`: if `hub.mode=subscribe` and `hub.verify_token` matches config, echo `hub.challenge`; otherwise 403.
- `POST /webhook/whatsapp`: compute HMAC-SHA256 of the **raw request body** with `WHATSAPP_APP_SECRET`, constant-time compare against `X-Hub-Signature-256` (format `sha256=<hex>`); reject with 401 on mismatch. If the app secret is unset, log a warning and skip verification (local-dev affordance).
- Persist the event by `wa_message_id` in `webhook_events` **before** processing; duplicate deliveries (Meta retries) are ignored via the primary key.
- Always return 200 quickly after processing; unknown senders (≠ `OWNER_PHONE_NUMBER`) are logged and ignored.
- The webhook route is exempt from the global 30 req/min rate limiter; `/api/*` keeps it.

**D6 — `/api/contact` failure semantics.**
Save to SQLite first; the endpoint returns 200 once saved. Email and WhatsApp notifications are attempted synchronously but best-effort — failures are logged (including the Cloud API error code for WhatsApp) and never fail the request. `wa_status` records the outcome (`sent` / `failed:<code>` / `skipped_no_config`) and `notify_msg_id` stores the returned message ID on success.

**D7 — Reply email.**
Reuse the existing SMTP service with a new `SendReplyEmail` that sends to the contact's email with a clear subject (e.g. `Re: Your message to <business>`). Sending failure is logged; the reply row is still recorded.

## Risks / Trade-offs

- [Owner doesn't quote-reply and multiple contacts exist] → Fallback (D2.3) may misattribute to the most recent contact. Mitigation: notification text explicitly shows `#<id>` and asks to swipe-reply; `match_method` recorded for audit.
- [24h window closed and no approved template yet] → Owner notification silently fails (logged, `wa_status=failed:131047`). Mitigation: setup step to approve a template; the forward email still goes out.
- [SMTP down when reply arrives] → Reply is recorded in `replies` but the contact never receives it; webhook still returns 200 so Meta will not redeliver. Mitigation: `webhook_events`/`replies` tables retain everything for manual re-send; acceptable for this scale.
- [Meta redelivers webhooks] → Deduped by `wa_message_id` primary key; safe.
- [SQLite write concurrency] → Single-process Gin server, low volume; WAL mode enabled as cheap insurance.
- [Owner's number changes phones/format] → `OWNER_PHONE_NUMBER` env comparison must normalize numbers (digits only, E.164 without `+` as Meta sends it). Mitigation: compare on normalized values.

## Migration Plan

1. Deploy is additive: new tables, new env vars, new route; existing `/api/contact` email behavior unchanged. No rollback migration needed (SQLite file can be deleted; only the new tables live in it).
2. Post-deploy manual steps: set env vars; configure callback URL `https://<domain>/webhook/whatsapp` + verify token in Meta App Dashboard and subscribe to the `messages` field; create and await approval of the utility template (D4); until then, owner sends one message to the business number to open the 24h window during testing.

## Open Questions

None blocking implementation. SMS forwarding and WhatsApp-messaging-the-contact are deferred (documented in proposal Non-Goals/Out of scope).
