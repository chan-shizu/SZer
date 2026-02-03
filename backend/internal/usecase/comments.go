package usecase

import (
	"context"
	"database/sql"

	"github.com/chan-shizu/SZer/db"
)

type CommentsUsecase struct {
	db *db.Queries
}

func NewCommentsUsecase(q *db.Queries) *CommentsUsecase {
	return &CommentsUsecase{db: q}
}

// コメント作成＆user_name付きで返す
func (u *CommentsUsecase) PostComment(ctx context.Context, programID int64, userID string, content string) (db.CreateCommentWithUserNameRow, error) {
	// 1. コメントINSERT
	_, err := u.db.CreateComment(ctx, db.CreateCommentParams{
		ProgramID: programID,
		UserID:    sqlNullString(userID),
		Content:   content,
	})
	if err != nil {
		return db.CreateCommentWithUserNameRow{}, err
	}
	// 2. user_name付きで返す
	return u.db.CreateCommentWithUserName(ctx, db.CreateCommentWithUserNameParams{
		ProgramID: programID,
		UserID:    sqlNullString(userID),
		Content:   content,
	})
}

func sqlNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}
