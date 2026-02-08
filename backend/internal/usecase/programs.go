package usecase

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"os"
	"strings"
	"time"

	"github.com/chan-shizu/SZer/db"
)

var ErrProgramNotFound = errors.New("program not found")
var ErrAlreadyPurchased = errors.New("already purchased")
var ErrInsufficientPoints = errors.New("insufficient points")
var ErrNotPurchasable = errors.New("program is not purchasable")

type ProgramDetailsCategoryTag struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type ProgramDetailsPerformer struct {
	ID           int64   `json:"id"`
	FullName     string  `json:"full_name"`
	FullNameKana string  `json:"full_name_kana"`
	ImageUrl     *string `json:"image_url"`
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
	IsLimitedRelease bool                        `json:"is_limited_release"`
	Price            int32                       `json:"price"`
	ThumbnailUrl     *string                     `json:"thumbnail_url"`
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
	IsLimitedRelease bool                        `json:"is_limited_release"`
	Price            int32                       `json:"price"`
	ThumbnailUrl     *string                     `json:"thumbnail_url"`
	CategoryTags     []ProgramDetailsCategoryTag `json:"category_tags"`
}

type TopProgramItem struct {
	ProgramID        int64   `json:"program_id"`
	Title            string  `json:"title"`
	ViewCount        int64   `json:"view_count"`
	LikeCount        int64   `json:"like_count"`
	IsLimitedRelease bool    `json:"is_limited_release"`
	Price            int32   `json:"price"`
	ThumbnailUrl     *string `json:"thumbnail_url"`
}

type ProgramsUsecase struct {
	conn *sql.DB
	q    *db.Queries
}

func NewProgramsUsecase(conn *sql.DB, q *db.Queries) *ProgramsUsecase {
	return &ProgramsUsecase{conn: conn, q: q}
}

func (u *ProgramsUsecase) UpsertWatchHistory(ctx context.Context, userID string, programID int64, positionSeconds int32, isCompleted bool) (db.WatchHistory, error) {
	_, err := u.q.GetIncompleteWatchHistoryByUserAndProgram(ctx, db.GetIncompleteWatchHistoryByUserAndProgramParams{
		UserID:    userID,
		ProgramID: programID,
	})
	log.Printf("[UpsertWatchHistory] userID=%s programID=%d positionSeconds=%d isCompleted=%v", userID, programID, positionSeconds, isCompleted)
	if err != nil {
		if err == sql.ErrNoRows {
			// 未完了履歴がなければINSERT
			log.Printf("[UpsertWatchHistory] 未完了履歴なし→INSERT")
			return u.q.InsertIncompleteWatchHistory(ctx, db.InsertIncompleteWatchHistoryParams{
				UserID:          userID,
				ProgramID:       programID,
				PositionSeconds: positionSeconds,
				IsCompleted:     isCompleted,
			})
		}
		log.Printf("[UpsertWatchHistory] 未完了履歴取得失敗: %v", err)
		return db.WatchHistory{}, err
	}
	// 未完了履歴があればUPDATE
	res, err := u.q.UpdateIncompleteWatchHistory(ctx, db.UpdateIncompleteWatchHistoryParams{
		UserID:          userID,
		ProgramID:       programID,
		PositionSeconds: positionSeconds,
		IsCompleted:     isCompleted,
	})
	
	if err != nil {
		log.Printf("[UpsertWatchHistory] UPDATE失敗: %v", err)
	} else {
		log.Printf("[UpsertWatchHistory] UPDATE成功: %+v", res)
	}

	return res, err
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
	if userID != "" {
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
			ImageUrl:     buildPublicFileURLPtr(p.ImagePath),
		})
	}

	resp := ProgramDetail{
		ProgramID:        program.ProgramID,
		Title:            program.Title,
		VideoURL:         buildVideoURL(program.VideoPath),
		ViewCount:        int64(program.ViewCount),
		LikeCount:        program.LikeCount,
		Liked:            program.Liked,
		IsLimitedRelease: program.IsLimitedRelease,
		Price:            program.Price,
		ThumbnailUrl:     buildPublicFileURLPtr(nullStringPtr(program.ThumbnailPath)),
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
			ProgramID:        program.ProgramID,
			Title:            program.Title,
			ViewCount:        int64(program.ViewCount),
			LikeCount:        program.LikeCount,
			IsLimitedRelease: program.IsLimitedRelease,
			Price:            program.Price,
			ThumbnailUrl:     buildPublicFileURLPtr(nullStringPtr(program.ThumbnailPath)),
			CategoryTags:     categoryTags,
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
			ProgramID:        program.ProgramID,
			Title:            program.Title,
			ViewCount:        int64(program.ViewCount),
			LikeCount:        program.LikeCount,
			IsLimitedRelease: program.IsLimitedRelease,
			Price:            program.Price,
			ThumbnailUrl:     buildPublicFileURLPtr(nullStringPtr(program.ThumbnailPath)),
		})
	}

	return results, nil
}

func (u *ProgramsUsecase) ListTopLikedPrograms(ctx context.Context) ([]TopProgramItem, error) {
	rows, err := u.q.GetTopLikedPrograms(ctx, sql.NullInt32{Int32: 7, Valid: true})
	if err != nil {
		return nil, err
	}

	results := make([]TopProgramItem, 0, len(rows))
	for _, row := range rows {
		results = append(results, TopProgramItem{
			ProgramID:        row.ProgramID,
			Title:            row.Title,
			ViewCount:        int64(row.ViewCount),
			LikeCount:        row.LikeCount,
			IsLimitedRelease: row.IsLimitedRelease,
			Price:            row.Price,
			ThumbnailUrl:     buildPublicFileURLPtr(nullStringPtr(row.ThumbnailPath)),
		})
	}

	return results, nil
}

func (u *ProgramsUsecase) ListTopViewedPrograms(ctx context.Context) ([]TopProgramItem, error) {
	rows, err := u.q.GetTopViewedPrograms(ctx, sql.NullInt32{Int32: 7, Valid: true})
	if err != nil {
		return nil, err
	}

	results := make([]TopProgramItem, 0, len(rows))
	for _, row := range rows {
		results = append(results, TopProgramItem{
			ProgramID:        row.ProgramID,
			Title:            row.Title,
			ViewCount:        int64(row.ViewCount),
			LikeCount:        row.LikeCount,
			IsLimitedRelease: row.IsLimitedRelease,
			Price:            row.Price,
			ThumbnailUrl:     buildPublicFileURLPtr(nullStringPtr(row.ThumbnailPath)),
		})
	}

	return results, nil
}

func (u *ProgramsUsecase) ListWatchingPrograms(ctx context.Context, userID string) ([]ProgramListItem, error) {
	rows, err := u.q.ListWatchingProgramsByUser(ctx, db.ListWatchingProgramsByUserParams{UserID: userID})
	if err != nil {
		return nil, err
	}

	results := make([]ProgramListItem, 0, len(rows))
	for _, row := range rows {
		categoryTagsJSON, err := normalizeJSONBytes(row.CategoryTags)
		if err != nil {
			return nil, err
		}
		var categoryTags []ProgramDetailsCategoryTag
		if err := json.Unmarshal(categoryTagsJSON, &categoryTags); err != nil {
			return nil, err
		}

		results = append(results, ProgramListItem{
			ProgramID:        row.ProgramID,
			Title:            row.Title,
			ViewCount:        int64(row.ViewCount),
			LikeCount:        row.LikeCount,
			IsLimitedRelease: row.IsLimitedRelease,
			Price:            row.Price,
			ThumbnailUrl:     buildPublicFileURLPtr(nullStringPtr(row.ThumbnailPath)),
			CategoryTags:     categoryTags,
		})
	}

	return results, nil
}

func (u *ProgramsUsecase) ListLikedPrograms(ctx context.Context, userID string) ([]ProgramListItem, error) {
	rows, err := u.q.ListLikedProgramsByUser(ctx, db.ListLikedProgramsByUserParams{UserID: userID})
	if err != nil {
		return nil, err
	}

	results := make([]ProgramListItem, 0, len(rows))
	for _, row := range rows {
		categoryTagsJSON, err := normalizeJSONBytes(row.CategoryTags)
		if err != nil {
			return nil, err
		}
		var categoryTags []ProgramDetailsCategoryTag
		if err := json.Unmarshal(categoryTagsJSON, &categoryTags); err != nil {
			return nil, err
		}

		results = append(results, ProgramListItem{
			ProgramID:        row.ProgramID,
			Title:            row.Title,
			ViewCount:        int64(row.ViewCount),
			LikeCount:        row.LikeCount,
			IsLimitedRelease: row.IsLimitedRelease,
			Price:            row.Price,
			ThumbnailUrl:     buildPublicFileURLPtr(nullStringPtr(row.ThumbnailPath)),
			CategoryTags:     categoryTags,
		})
	}

	return results, nil
}

func (u *ProgramsUsecase) ListPurchasedPrograms(ctx context.Context, userID string) ([]ProgramListItem, error) {
	rows, err := u.q.ListPurchasedProgramsByUser(ctx, db.ListPurchasedProgramsByUserParams{UserID: userID})
	if err != nil {
		return nil, err
	}

	results := make([]ProgramListItem, 0, len(rows))
	for _, row := range rows {
		categoryTagsJSON, err := normalizeJSONBytes(row.CategoryTags)
		if err != nil {
			return nil, err
		}
		var categoryTags []ProgramDetailsCategoryTag
		if err := json.Unmarshal(categoryTagsJSON, &categoryTags); err != nil {
			return nil, err
		}

		results = append(results, ProgramListItem{
			ProgramID:        row.ProgramID,
			Title:            row.Title,
			ViewCount:        int64(row.ViewCount),
			LikeCount:        row.LikeCount,
			IsLimitedRelease: row.IsLimitedRelease,
			Price:            row.Price,
			ThumbnailUrl:     buildPublicFileURLPtr(nullStringPtr(row.ThumbnailPath)),
			CategoryTags:     categoryTags,
		})
	}

	return results, nil
}

// 視聴回数をインクリメントするメソッドを追加
func (u *ProgramsUsecase) IncrementViewCount(ctx context.Context, programID int64) error {
	return u.q.IncrementProgramViewCount(ctx, programID)
}

// ポイントで番組を購入し、閲覧権限を付与する
func (u *ProgramsUsecase) PurchaseProgram(ctx context.Context, userID string, programID int64) (int32, error) {
	// 番組情報取得
	program, err := u.q.GetProgramForPurchase(ctx, programID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrProgramNotFound
		}
		return 0, err
	}

	// 購入可能かチェック（限定公開かつ有料のみ）
	if !program.IsLimitedRelease || program.Price <= 0 {
		return 0, ErrNotPurchasable
	}

	// 既に購入済みかチェック
	permitted, err := u.q.IsUserPermittedForProgram(ctx, db.IsUserPermittedForProgramParams{
		UserID:    userID,
		ProgramID: programID,
	})
	if err != nil {
		return 0, err
	}
	if permitted {
		return 0, ErrAlreadyPurchased
	}

	// トランザクション開始
	tx, err := u.conn.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()

	qtx := u.q.WithTx(tx)

	// ポイント差し引き（WHERE points >= price でアトミックにチェック）
	newPoints, err := qtx.DeductPointsFromUser(ctx, db.DeductPointsFromUserParams{
		ID:     userID,
		Points: program.Price,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrInsufficientPoints
		}
		return 0, err
	}

	// 閲覧権限付与
	if err := qtx.AddPermittedProgramUser(ctx, db.AddPermittedProgramUserParams{
		UserID:    userID,
		ProgramID: programID,
	}); err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return newPoints, nil
}

// 限定公開動画の閲覧権限チェック
func (u *ProgramsUsecase) IsUserPermittedForProgram(ctx context.Context, userID string, programID int64) (bool, error) {
	if userID == "" {
		return false, nil
	}
	return u.q.IsUserPermittedForProgram(ctx, db.IsUserPermittedForProgramParams{
		UserID:    userID,
		ProgramID: programID,
	})
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