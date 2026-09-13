package store

import (
	"database/sql"
	"errors"
	"fmt"

	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

type Contact struct {
	ID          int64
	Name        string
	Email       string
	Phone       string
	Message     string
	NotifyMsgID string
	WAStatus    string
	CreatedAt   string
}

const schema = `
CREATE TABLE IF NOT EXISTS contacts (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	email TEXT NOT NULL,
	phone TEXT NOT NULL,
	message TEXT NOT NULL,
	notify_msg_id TEXT,
	wa_status TEXT,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE IF NOT EXISTS replies (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	contact_id INTEGER NOT NULL REFERENCES contacts(id),
	wa_message_id TEXT UNIQUE NOT NULL,
	body TEXT NOT NULL,
	match_method TEXT NOT NULL,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE IF NOT EXISTS webhook_events (
	wa_message_id TEXT PRIMARY KEY,
	raw TEXT NOT NULL,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
`

func Open(path string) (*Store, error) {
	dsn := fmt.Sprintf("file:%s?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)", path)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate sqlite schema: %w", err)
	}

	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) CreateContact(name, email, phone, message string) (int64, error) {
	res, err := s.db.Exec(
		"INSERT INTO contacts (name, email, phone, message) VALUES (?, ?, ?, ?)",
		name, email, phone, message,
	)
	if err != nil {
		return 0, fmt.Errorf("insert contact: %w", err)
	}
	return res.LastInsertId()
}

func (s *Store) UpdateContactWhatsAppResult(id int64, notifyMsgID, status string) error {
	_, err := s.db.Exec(
		"UPDATE contacts SET notify_msg_id = ?, wa_status = ? WHERE id = ?",
		notifyMsgID, status, id,
	)
	if err != nil {
		return fmt.Errorf("update contact %d: %w", id, err)
	}
	return nil
}

func (s *Store) GetContactByNotifyMsgID(msgID string) (*Contact, error) {
	return s.queryContact("SELECT id, name, email, phone, message, notify_msg_id, wa_status, created_at FROM contacts WHERE notify_msg_id = ?", msgID)
}

func (s *Store) GetContactByID(id int64) (*Contact, error) {
	return s.queryContact("SELECT id, name, email, phone, message, notify_msg_id, wa_status, created_at FROM contacts WHERE id = ?", id)
}

func (s *Store) LatestContact() (*Contact, error) {
	return s.queryContact("SELECT id, name, email, phone, message, notify_msg_id, wa_status, created_at FROM contacts ORDER BY id DESC LIMIT 1")
}

func (s *Store) queryContact(query string, args ...any) (*Contact, error) {
	row := s.db.QueryRow(query, args...)
	var c Contact
	err := row.Scan(&c.ID, &c.Name, &c.Email, &c.Phone, &c.Message, &c.NotifyMsgID, &c.WAStatus, &c.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query contact: %w", err)
	}
	return &c, nil
}

func (s *Store) InsertWebhookEvent(waMessageID, raw string) (bool, error) {
	res, err := s.db.Exec(
		"INSERT OR IGNORE INTO webhook_events (wa_message_id, raw) VALUES (?, ?)",
		waMessageID, raw,
	)
	if err != nil {
		return false, fmt.Errorf("insert webhook event: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("webhook event rows affected: %w", err)
	}
	return n == 1, nil
}

func (s *Store) InsertReply(contactID int64, waMessageID, body, matchMethod string) error {
	_, err := s.db.Exec(
		"INSERT INTO replies (contact_id, wa_message_id, body, match_method) VALUES (?, ?, ?, ?)",
		contactID, waMessageID, body, matchMethod,
	)
	if err != nil {
		return fmt.Errorf("insert reply: %w", err)
	}
	return nil
}
