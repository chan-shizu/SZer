package usecase

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/chan-shizu/SZer/db"
)

var ErrRequestContentRequired = errors.New("content is required")
var ErrRequestNameRequired = errors.New("name is required")
var ErrRequestContactRequired = errors.New("contact is required")

type RequestsUsecase struct {
	q *db.Queries
}

func NewRequestsUsecase(q *db.Queries) *RequestsUsecase {
	return &RequestsUsecase{q: q}
}

func (u *RequestsUsecase) CreateRequest(ctx context.Context, userID string, content, name, contact, note string) (db.Request, error) {
	content = strings.TrimSpace(content)
	name = strings.TrimSpace(name)
	contact = strings.TrimSpace(contact)
	note = strings.TrimSpace(note)

	if content == "" {
		return db.Request{}, ErrRequestContentRequired
	}
	if name == "" {
		return db.Request{}, ErrRequestNameRequired
	}
	if contact == "" {
		return db.Request{}, ErrRequestContactRequired
	}

	userIDNull := sql.NullString{String: userID, Valid: userID != ""}

	return u.q.CreateRequest(ctx, db.CreateRequestParams{
		UserID:  userIDNull,
		Content: content,
		Name:    name,
		Contact: contact,
		Note:    note,
	})
}
