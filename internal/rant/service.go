package rant

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Store interface {
	Save(ctx context.Context, record EncryptedRecord) error
	FindAll(ctx context.Context) ([]EncryptedRecord, error)
	Wipe() error
}

type Cipher interface {
	Encrypt(plaintext []byte) (ciphertext []byte, nonce []byte, err error)
	Decrypt(ciphertext []byte, nonce []byte) ([]byte, error)
}

type KeyStore interface {
	DeleteMasterKey() error
}

type RecordInput struct {
	Content  string
	Severity int
	Tags     []string
}

type Service struct {
	store    Store
	cipher   Cipher
	keyStore KeyStore
}

func NewService(store Store, cipher Cipher, keyStore KeyStore) *Service {
	return &Service{
		store:    store,
		cipher:   cipher,
		keyStore: keyStore,
	}
}

func (s *Service) Record(ctx context.Context, input RecordInput) (Rant, error) {
	severity, err := ParseSeverity(input.Severity)
	if err != nil {
		return Rant{}, err
	}

	trimmedContent := strings.TrimSpace(input.Content)
	if trimmedContent == "" {
		return Rant{}, ErrEmptyContent
	}

	id, err := uuid.NewV7()
	if err != nil {
		return Rant{}, err
	}

	encryptedPayload, nonce, err := s.cipher.Encrypt([]byte(trimmedContent))
	if err != nil {
		return Rant{}, err
	}

	now := time.Now()
	tags := normalizeTags(input.Tags)

	record := EncryptedRecord{
		ID:               id.String(),
		CreatedAt:        now.UTC(),
		DayOfWeek:        int(now.Weekday()),
		HourOfDay:        now.Hour(),
		Severity:         severity,
		EncryptedPayload: encryptedPayload,
		Nonce:            nonce,
		Tags:             tags,
	}

	if err := s.store.Save(ctx, record); err != nil {
		return Rant{}, err
	}

	return Rant{
		ID:        record.ID,
		CreatedAt: record.CreatedAt,
		Severity:  record.Severity,
		Content:   trimmedContent,
		Tags:      record.Tags,
	}, nil
}

func (s *Service) List(ctx context.Context) ([]Rant, error) {
	records, err := s.store.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	rants := make([]Rant, 0, len(records))
	for _, rec := range records {
		plaintext, err := s.cipher.Decrypt(rec.EncryptedPayload, rec.Nonce)
		if err != nil {
			return nil, err
		}

		rants = append(rants, Rant{
			ID:        rec.ID,
			CreatedAt: rec.CreatedAt,
			Severity:  rec.Severity,
			Content:   string(plaintext),
			Tags:      rec.Tags,
		})
	}

	return rants, nil
}

func (s *Service) Nuke() error {
	if err := s.store.Wipe(); err != nil {
		return err
	}

	return s.keyStore.DeleteMasterKey()
}
