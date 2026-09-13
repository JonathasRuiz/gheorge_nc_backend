## ADDED Requirements

### Requirement: Notify owner via WhatsApp on contact submission
The system SHALL send a WhatsApp message from the business number to the configured OWNER_PHONE_NUMBER for each persisted contact submission, containing the contact id tag (#<id>), name, and message, so the owner can identify and reply to the inquiry.

#### Scenario: Submission triggers WhatsApp notification
- **WHEN** a contact form submission is saved
- **THEN** a WhatsApp message containing "#<id>", the contact name, and the message text is sent to the owner's number

### Requirement: Template-aware sending
The system SHALL use the configured WHATSAPP_TEMPLATE_NAME template (parameters: id tag, name, message) when it is set, and SHALL send free-form text when it is not.

#### Scenario: Template configured
- **WHEN** WHATSAPP_TEMPLATE_NAME is set
- **THEN** the notification is sent as that template with the id tag, name, and message as parameters

#### Scenario: No template configured
- **WHEN** WHATSAPP_TEMPLATE_NAME is empty
- **THEN** the notification is sent as free-form text

### Requirement: Notification is best-effort
The system SHALL NOT fail the contact form request due to notification failures: both the email forward to CONTACT_FORWARD_TO and the WhatsApp notification are attempted after the save, and any failure SHALL be logged with its error detail while the endpoint still returns HTTP 200.

#### Scenario: WhatsApp failure does not fail the form
- **WHEN** the Cloud API rejects the owner notification (e.g. 24h window closed, error 131047)
- **THEN** the failure and error code are logged, the contact row records the failure, the forward email is still attempted, and the API responds HTTP 200

#### Scenario: No WhatsApp credentials configured
- **WHEN** WHATSAPP_TOKEN or WHATSAPP_PHONE_NUMBER_ID or OWNER_PHONE_NUMBER is empty
- **THEN** the WhatsApp notification is skipped, logged, and the submission still succeeds with HTTP 200

### Requirement: Keep email notification
The system SHALL retain the existing behavior of emailing the submission to CONTACT_FORWARD_TO in addition to the WhatsApp notification.

#### Scenario: Both channels notified
- **WHEN** a contact submission is saved and both channels are configured
- **THEN** the forward email is sent and the WhatsApp notification is sent
