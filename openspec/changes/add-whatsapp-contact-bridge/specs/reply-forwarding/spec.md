## ADDED Requirements

### Requirement: Correlate owner replies to contacts
The system SHALL treat every inbound WhatsApp message from the owner's number as a reply to a contact, resolving the target contact with a deterministic cascade: (1) the quoted context.message_id matches a stored contacts.notify_msg_id; (2) the message body begins with a "#<id>" tag matching a stored contact id; (3) otherwise the most recently created contact. The matched method SHALL be recorded with the reply.

#### Scenario: Owner swipe-replies to a notification
- **WHEN** the owner uses WhatsApp's reply feature on a notification message and the webhook payload contains context.message_id equal to a stored notify_msg_id
- **THEN** the reply is correlated to that contact with match_method quote

#### Scenario: Owner prefixes the id tag
- **WHEN** the owner's message body starts with "#42"
- **THEN** the reply is correlated to contact id 42 with match_method tag

#### Scenario: Owner sends an untagged message
- **WHEN** the owner's message has no context.message_id and no "#<id>" prefix
- **THEN** the reply is correlated to the most recently created contact with match_method fallback

#### Scenario: No contacts exist
- **WHEN** the owner sends a message and the contacts table is empty
- **THEN** no correlation occurs, the event is logged, and no email is sent

### Requirement: Forward reply to contact email
The system SHALL forward the owner's reply text to the correlated contact's email address via the existing SMTP configuration with a reply-style subject, and SHALL record the forwarded reply (contact id, WhatsApp message id, body, match method) in the replies table.

#### Scenario: Reply forwarded
- **WHEN** an owner reply is correlated to a contact
- **THEN** an email containing the owner's reply text is sent to the contact's email address and a replies row is inserted

#### Scenario: Forwarding failure is recorded not retried
- **WHEN** the SMTP send fails
- **THEN** the failure is logged, the reply row is still recorded, and the webhook response remains HTTP 200

### Requirement: Reply deduplication
The system SHALL NOT forward the same owner WhatsApp message twice, regardless of webhook redelivery.

#### Scenario: Same reply delivered twice by Meta
- **WHEN** Meta redelivers an owner message webhook
- **THEN** the contact receives the forwarded email exactly once
