## ADDED Requirements

### Requirement: Persist contact form submissions
The system SHALL persist every valid contact form submission (name, email, phone, message) to the SQLite database before attempting any notification, and the submission SHALL be assigned a unique incremental id.

#### Scenario: Valid submission is stored
- **WHEN** a POST /api/contact request contains valid name, email, phone, and message
- **THEN** the system inserts a row into the contacts table and returns HTTP 200

#### Scenario: Storage precedes notification
- **WHEN** the contact submission is saved
- **THEN** the save occurs before any email or WhatsApp notification is attempted, and a subsequent notification failure does not remove or alter the stored row

### Requirement: Record WhatsApp notification outcome on the contact
The system SHALL record on the contact row the outcome of the WhatsApp notification attempt: the Cloud API message ID on success, or a failure/skipped status with the error code.

#### Scenario: WhatsApp send succeeds
- **WHEN** the Cloud API accepts the notification message
- **THEN** the returned message ID is stored in contacts.notify_msg_id and wa_status is set to sent

#### Scenario: WhatsApp send fails
- **WHEN** the Cloud API rejects the notification message
- **THEN** wa_status is set to failed with the Cloud API error code and notify_msg_id remains empty

### Requirement: Database initialization
The system SHALL create the SQLite database file and all required tables (contacts, replies, webhook_events) automatically on startup using the DB_PATH environment variable (default: data.db in the working directory), and existing data SHALL be preserved across restarts.

#### Scenario: Fresh start creates schema
- **WHEN** the server starts and the database file does not exist
- **THEN** the file and all tables are created with CREATE TABLE IF NOT EXISTS semantics

#### Scenario: Restart preserves data
- **WHEN** the server restarts with an existing database file
- **THEN** previously stored contacts, replies, and webhook events remain intact
