package usecase

import (
	"context"

	"github.com/chan-shizu/SZer/db"
)

func (u *CommentsUsecase) ListCommentsByProgramID(ctx context.Context, programID int64) ([]db.ListCommentsByProgramIDRow, error) {
	return u.db.ListCommentsByProgramID(ctx, programID)
}
