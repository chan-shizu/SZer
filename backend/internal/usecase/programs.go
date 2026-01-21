package usecase

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"time"

	"github.com/chan-shizu/SZer/db"
)

var ErrProgramNotFound = errors.New("program not found")

type ProgramDetailsCategoryTag struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type ProgramDetailsPerformer struct {
	ID           int64   `json:"id"`
	FullName     string  `json:"full_name"`
	FullNameKana string  `json:"full_name_kana"`
	ImageUrl    *string `json:"image_url"`
}

type ProgramWatchHistory struct {
	PositionSeconds int32     `json:"position_seconds"`
	IsCompleted     bool      `json:"is_completed"`
	LastWatchedAt   time.Time `json:"last_watched_at"`
}

type programDetailsPerformerRaw struct {
	ID            int64   `json:"id"`
	FirstName     string  `json:"first_name"`
	LastName      string  `json:"last_name"`
	FirstNameKana string  `json:"first_name_kana"`
	LastNameKana  string  `json:"last_name_kana"`
	ImagePath     *string `json:"image_path"`
}

type ProgramDetail struct {
	ProgramID        int64                       `json:"program_id"`
	Title            string                      `json:"title"`
	VideoURL         string                      `json:"video_url"`
	ViewCount        int64                       `json:"view_count"`
	LikeCount        int64                       `json:"like_count"`
	Liked            bool                        `json:"liked"`
	ThumbnailUrl    *string                     `json:"thumbnail_url"`
	Description      *string                     `json:"description"`
	ProgramCreatedAt time.Time                   `json:"program_created_at"`
	ProgramUpdatedAt time.Time                   `json:"program_updated_at"`
	CategoryTags     []ProgramDetailsCategoryTag `json:"category_tags"`
	Performers       []ProgramDetailsPerformer   `json:"performers"`
	WatchHistory     *ProgramWatchHistory        `json:"watch_history"`
}

type ProgramListItem struct {
	ProgramID        int64                       `json:"program_id"`
	Title            string                      `json:"title"`
	ViewCount        int64                       `json:"view_count"`
	LikeCount        int64                       `json:"like_count"`
	ThumbnailUrl    *string                     `json:"thumbnail_url"`
	CategoryTags     []ProgramDetailsCategoryTag `json:"category_tags"`
}

type TopProgramItem struct {
	ProgramID     int64   `json:"program_id"`
	Title         string  `json:"title"`
	ViewCount     int64   `json:"view_count"`
	LikeCount     int64   `json:"like_count"`
	ThumbnailUrl  *string `json:"thumbnail_url"`
}

type ProgramsUsecase struct {
	q *db.Queries
}

func NewProgramsUsecase(q *db.Queries) *ProgramsUsecase {
	return &ProgramsUsecase{q: q}
}

func (u *ProgramsUsecase) UpsertWatchHistory(ctx context.Context, userID string, programID int64, positionSeconds int32, isCompleted bool) (db.UpsertWatchHistoryRow, error) {
	return u.q.UpsertWatchHistory(ctx, db.UpsertWatchHistoryParams{
		UserID:          userID,
		ProgramID:       programID,
		PositionSeconds: positionSeconds,
		IsCompleted:     isCompleted,
	})
}

func (u *ProgramsUsecase) GetProgramDetails(ctx context.Context, userID string, id int64) (ProgramDetail, error) {
	program, err := u.q.GetProgramDetailsByID(ctx, db.GetProgramDetailsByIDParams{ID: id, UserID: userID})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ProgramDetail{}, ErrProgramNotFound
		}
		return ProgramDetail{}, err
	}

	var watchHistory *ProgramWatchHistory
	wh, err := u.q.GetIncompleteWatchHistoryByUserAndProgram(ctx, db.GetIncompleteWatchHistoryByUserAndProgramParams{UserID: userID, ProgramID: id})
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return ProgramDetail{}, err
		}
	} else {
		watchHistory = &ProgramWatchHistory{
			PositionSeconds: wh.PositionSeconds,
			IsCompleted:     wh.IsCompleted,
			LastWatchedAt:   wh.LastWatchedAt,
		}
	}

	categoryTagsJSON, err := normalizeJSONBytes(program.CategoryTags)
	if err != nil {
		return ProgramDetail{}, err
	}
	performersJSON, err := normalizeJSONBytes(program.Performers)
	if err != nil {
		return ProgramDetail{}, err
	}

	var categoryTags []ProgramDetailsCategoryTag
	if err := json.Unmarshal(categoryTagsJSON, &categoryTags); err != nil {
		return ProgramDetail{}, err
	}

	var performersRaw []programDetailsPerformerRaw
	if err := json.Unmarshal(performersJSON, &performersRaw); err != nil {
		return ProgramDetail{}, err
	}

	performers := make([]ProgramDetailsPerformer, 0, len(performersRaw))
	for _, p := range performersRaw {
		performers = append(performers, ProgramDetailsPerformer{
			ID:           p.ID,
			FullName:     p.LastName + p.FirstName,
			FullNameKana: p.LastNameKana + p.FirstNameKana,
			ImageUrl:    buildPublicFileURLPtr(p.ImagePath),
		})
	}

	resp := ProgramDetail{
		ProgramID:        program.ProgramID,
		Title:            program.Title,
		VideoURL:         buildVideoURL(program.VideoPath),
		ViewCount:        program.ViewCount,
		LikeCount:        program.LikeCount,
		Liked:            program.Liked,
		ThumbnailUrl:    buildPublicFileURLPtr(nullStringPtr(program.ThumbnailPath)),
		Description:      nullStringPtr(program.Description),
		ProgramCreatedAt: program.ProgramCreatedAt,
		ProgramUpdatedAt: program.ProgramUpdatedAt,
		CategoryTags:     categoryTags,
		Performers:       performers,
		WatchHistory:     watchHistory,
	}
	return resp, nil
}

func (u *ProgramsUsecase) LikeProgram(ctx context.Context, userID string, programID int64) (bool, int64, error) {
	exists, err := u.q.ExistsProgram(ctx, programID)
	if err != nil {
		return false, 0, err
	}
	if !exists {
		return false, 0, ErrProgramNotFound
	}

	if err := u.q.CreateLike(ctx, db.CreateLikeParams{UserID: userID, ProgramID: programID}); err != nil {
		return false, 0, err
	}

	likeCount, err := u.q.CountLikesByProgramID(ctx, programID)
	if err != nil {
		return false, 0, err
	}
	return true, likeCount, nil
}

func (u *ProgramsUsecase) UnlikeProgram(ctx context.Context, userID string, programID int64) (bool, int64, error) {
	exists, err := u.q.ExistsProgram(ctx, programID)
	if err != nil {
		return false, 0, err
	}
	if !exists {
		return false, 0, ErrProgramNotFound
	}

	if err := u.q.DeleteLike(ctx, db.DeleteLikeParams{UserID: userID, ProgramID: programID}); err != nil {
		return false, 0, err
	}

	likeCount, err := u.q.CountLikesByProgramID(ctx, programID)
	if err != nil {
		return false, 0, err
	}
	return false, likeCount, nil
}

func (u *ProgramsUsecase) ListPrograms(ctx context.Context, title string, tagIDs []int64) ([]ProgramListItem, error) {
	arg := db.GetProgramsParams{}
	if title != "" {
		arg.Title = sql.NullString{String: title, Valid: true}
	}
	if len(tagIDs) > 0 {
		arg.TagIds = tagIDs
	}

	programs, err := u.q.GetPrograms(ctx, arg)
	if err != nil {
		return nil, err
	}

	var results []ProgramListItem
	for _, program := range programs {
		categoryTagsJSON, err := normalizeJSONBytes(program.CategoryTags)
		if err != nil {
			return nil, err
		}
		var categoryTags []ProgramDetailsCategoryTag
		if err := json.Unmarshal(categoryTagsJSON, &categoryTags); err != nil {
			return nil, err // またはログを出して空配列にするなど、エラーハンドリングポリシーによる
		}

		results = append(results, ProgramListItem{
			ProgramID:     program.ProgramID,
			Title:         program.Title,
			ViewCount:     program.ViewCount,
			LikeCount:     program.LikeCount,
			ThumbnailUrl:  buildPublicFileURLPtr(nullStringPtr(program.ThumbnailPath)),
			CategoryTags:  categoryTags,
		})
	}

	return results, nil
}

func (u *ProgramsUsecase) ListTopPrograms(ctx context.Context) ([]TopProgramItem, error) {
	programs, err := u.q.GetTopPrograms(ctx)
	if err != nil {
		return nil, err
	}

	results := make([]TopProgramItem, 0, len(programs))
	for _, program := range programs {
		results = append(results, TopProgramItem{
			ProgramID:    program.ProgramID,
			Title:        program.Title,
			ViewCount:    program.ViewCount,
			LikeCount:    program.LikeCount,
			ThumbnailUrl: buildPublicFileURLPtr(nullStringPtr(program.ThumbnailPath)),
		})
	}

	return results, nil
}


// private functions

func buildVideoURL(videoPath string) string {
	if strings.HasPrefix(videoPath, "http://") || strings.HasPrefix(videoPath, "https://") {
		return videoPath
	}

	base := os.Getenv("S3_VIDEO_BUCKET_ENDPOINT")
	if base == "" {
		return videoPath
	}

	base = strings.TrimRight(base, "/")
	path := strings.TrimLeft(videoPath, "/")
	return base + "/" + path
}

func nullStringPtr(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	v := ns.String
	return &v
}

func buildPublicFileURL(filePath string) string {
	if filePath == "" {
		return ""
	}
	if strings.HasPrefix(filePath, "http://") || strings.HasPrefix(filePath, "https://") {
		return filePath
	}

	base := os.Getenv("S3_PUBLIC_FILE_BUCKET_ENDPOINT")
	if base == "" {
		return filePath
	}

	base = strings.TrimRight(base, "/")
	path := strings.TrimLeft(filePath, "/")
	return base + "/" + path
}

func buildPublicFileURLPtr(filePath *string) *string {
	if filePath == nil {
		return nil
	}
	v := buildPublicFileURL(*filePath)
	if v == "" {
		return nil
	}
	return &v
}

func normalizeJSONBytes(v interface{}) ([]byte, error) {
	switch vv := v.(type) {
	case nil:
		return []byte("[]"), nil
	case []byte:
		return vv, nil
	case string:
		return []byte(vv), nil
	default:
		b, err := json.Marshal(vv)
		if err != nil {
			return nil, err
		}
		return b, nil
	}
}
