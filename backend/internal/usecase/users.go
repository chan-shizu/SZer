package usecase

import (
	"context"
	"database/sql"
	"errors"

	"github.com/chan-shizu/SZer/db"
)

var ErrUserNotFound = errors.New("user not found")
var ErrInvalidPointsAmount = errors.New("invalid points amount")

type UsersUsecase struct {
	q *db.Queries
}

func NewUsersUsecase(q *db.Queries) *UsersUsecase {
	return &UsersUsecase{q: q}
}

func (u *UsersUsecase) AddPoints(ctx context.Context, userID string, amount int32) (int32, error) {
	switch amount {
	case 100, 500, 1000:
		// ok
	default:
		return 0, ErrInvalidPointsAmount
	}

	points, err := u.q.AddPointsToUser(ctx, db.AddPointsToUserParams{ID: userID, Points: amount})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrUserNotFound
		}
		return 0, err
	}
	return points, nil
}

func (u *UsersUsecase) GetPoints(ctx context.Context, userID string) (int32, error) {
	points, err := u.q.GetUserPoints(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrUserNotFound
		}
		return 0, err
	}
	return points, nil
}
