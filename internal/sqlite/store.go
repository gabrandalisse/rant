package sqlite

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gabrandalisse/rant/internal/rant"
	_ "modernc.org/sqlite"
)

const schema = `
CREATE TABLE IF NOT EXISTS rants (
	id TEXT PRIMARY KEY,
	created_at INTEGER NOT NULL,
	day_of_week INTEGER NOT NULL,
	hour_of_day INTEGER NOT NULL,
	severity INTEGER NOT NULL,
	encrypted_payload BLOB NOT NULL,
	nonce BLOB NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_rants_created_at ON rants(created_at);

CREATE TABLE IF NOT EXISTS rant_tags (
	rant_id TEXT NOT NULL REFERENCES rants(id) ON DELETE CASCADE,
	tag TEXT NOT NULL,
	PRIMARY KEY (rant_id, tag)
);

CREATE INDEX IF NOT EXISTS idx_rant_tags_tag ON rant_tags(tag);
`

type Store struct {
	db     *sql.DB
	dbPath string
}

func Open(dbPath string) (*Store, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("create vault directory: %w", err)
	}

	dsn := fmt.Sprintf("file:%s?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)", dbPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}

	if err := os.Chmod(dbPath, 0o600); err != nil && !os.IsNotExist(err) {
		_ = db.Close()
		return nil, fmt.Errorf("enforce file permissions: %w", err)
	}

	if _, err := db.Exec(schema); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("apply schema: %w", err)
	}

	return &Store{
		db:     db,
		dbPath: dbPath,
	}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) Save(ctx context.Context, record rant.EncryptedRecord) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() { _ = tx.Rollback() }()

	_, err = tx.ExecContext(
		ctx,
		`INSERT INTO rants (id, created_at, day_of_week, hour_of_day, severity, encrypted_payload, nonce)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		record.ID,
		record.CreatedAt.Unix(),
		record.DayOfWeek,
		record.HourOfDay,
		int(record.Severity),
		record.EncryptedPayload,
		record.Nonce,
	)
	if err != nil {
		return err
	}

	for _, tag := range record.Tags {
		_, err = tx.ExecContext(
			ctx,
			`INSERT INTO rant_tags (rant_id, tag) VALUES (?, ?)`,
			record.ID,
			tag,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *Store) FindAll(ctx context.Context) ([]rant.EncryptedRecord, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT id, created_at, day_of_week, hour_of_day, severity, encrypted_payload, nonce
		 FROM rants
		 ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var records []rant.EncryptedRecord
	recordIndex := make(map[string]int)

	for rows.Next() {
		var rec rant.EncryptedRecord
		var createdAtUnix int64
		var severityInt int

		err := rows.Scan(
			&rec.ID,
			&createdAtUnix,
			&rec.DayOfWeek,
			&rec.HourOfDay,
			&severityInt,
			&rec.EncryptedPayload,
			&rec.Nonce,
		)
		if err != nil {
			return nil, err
		}

		rec.CreatedAt = time.Unix(createdAtUnix, 0).UTC()
		rec.Severity = rant.Severity(severityInt)
		rec.Tags = []string{}

		recordIndex[rec.ID] = len(records)
		records = append(records, rec)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(records) == 0 {
		return records, nil
	}

	tagRows, err := s.db.QueryContext(
		ctx,
		`SELECT rant_id, tag FROM rant_tags`,
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tagRows.Close() }()

	for tagRows.Next() {
		var rantID, tag string
		if err := tagRows.Scan(&rantID, &tag); err != nil {
			return nil, err
		}
		if idx, exists := recordIndex[rantID]; exists {
			records[idx].Tags = append(records[idx].Tags, tag)
		}
	}

	return records, tagRows.Err()
}

func (s *Store) Wipe() error {
	_ = s.db.Close()

	targets := []string{
		s.dbPath,
		s.dbPath + "-wal",
		s.dbPath + "-shm",
	}

	for _, target := range targets {
		if err := overwriteAndRemove(target); err != nil {
			return err
		}
	}

	return nil
}

func overwriteAndRemove(path string) error {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil
	}

	if err != nil {
		return err
	}

	file, err := os.OpenFile(path, os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}

	noise := make([]byte, info.Size())
	if _, err := rand.Read(noise); err != nil {
		_ = file.Close()
		return err
	}

	if _, err := file.WriteAt(noise, 0); err != nil {
		_ = file.Close()
		return err
	}

	if err := file.Sync(); err != nil {
		_ = file.Close()
		return err
	}

	if err := file.Close(); err != nil {
		return err
	}

	return os.Remove(path)
}
