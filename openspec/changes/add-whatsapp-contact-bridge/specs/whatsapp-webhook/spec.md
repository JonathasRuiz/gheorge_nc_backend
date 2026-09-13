## ADDED Requirements

### Requirement: Meta webhook verification handshake
The system SHALL expose GET /webhook/whatsapp that, when hub.mode equals "subscribe" and hub.verify_token matches the configured WHATSAPP_VERIFY_TOKEN, responds with the hub.challenge value; all other cases SHALL return HTTP 403.

#### Scenario: Valid verification request
- **WHEN** Meta sends GET /webhook/whatsapp with hub.mode=subscribe and the correct hub.verify_token
- **THEN** the system responds HTTP 200 with the hub.challenge value as the body

#### Scenario: Wrong verify token
- **WHEN** the hub.verify_token does not match the configured value
- **THEN** the system responds HTTP 403

### Requirement: Webhook payload signature verification
The system SHALL verify the X-Hub-Signature-256 header of POST /webhook/whatsapp by computing HMAC-SHA256 of the raw request body with WHATSAPP_APP_SECRET and comparing it in constant time; a mismatch SHALL be rejected with HTTP 401 and not processed. When WHATSAPP_APP_SECRET is unset, the system SHALL log a warning and process without verification.

#### Scenario: Valid signature
- **WHEN** a POST arrives whose X-Hub-Signature-256 equals sha256=<HMAC of raw body>
- **THEN** the payload is processed

#### Scenario: Tampered body
- **WHEN** the signature does not match the raw body
- **THEN** the system responds HTTP 401 and no database rows are written

#### Scenario: No app secret configured
- **WHEN** WHATSAPP_APP_SECRET is empty
- **THEN** the system logs a warning, skips verification, and processes the payload

### Requirement: Idempotent webhook processing
The system SHALL persist each inbound message event by its WhatsApp message ID (wamid) in the webhook_events table before processing, and any redelivered event with an already-stored message ID SHALL be ignored without side effects.

#### Scenario: Duplicate delivery
- **WHEN** the same webhook event is delivered twice
- **THEN** only the first delivery is processed and the second is ignored, with both returning HTTP 200

### Requirement: Fast webhook acknowledgment
The system SHALL respond to POST /webhook/whatsapp with HTTP 200 after processing regardless of processing outcome, so Meta does not retry unnecessarily.

#### Scenario: Processing error does not fail the response
- **WHEN** correlation or email forwarding fails during processing
- **THEN** the event remains recorded in the database and the response is still HTTP 200

### Requirement: Webhook rate-limit exemption
The system SHALL exempt /webhook/whatsapp from the global API rate limiter; /api routes SHALL keep the existing rate limiting.

#### Scenario: Webhook burst is not throttled
- **WHEN** more than 30 webhook POSTs arrive within one minute
- **THEN** all webhook requests are processed without rate-limit rejection

### Requirement: Ignore non-owner inbound messages
The system SHALL ignore inbound WhatsApp messages whose sender does not match the configured OWNER_PHONE_NUMBER (normalized comparison), logging them without correlation or forwarding.

#### Scenario: Public message to the business number
- **WHEN** a WhatsApp user other than the owner messages the business number
- **THEN** the event is recorded, no contact correlation occurs, and no email is sent
