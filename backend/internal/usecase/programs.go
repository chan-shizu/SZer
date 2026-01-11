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

type ProgramDetails struct {
	ProgramID        int64                       `json:"program_id"`
	Title            string                      `json:"title"`
	VideoURL         string                      `json:"video_url"`
	ThumbnailUrl    *string                     `json:"thumbnail_url"`
	Description      *string                     `json:"description"`
	ProgramCreatedAt time.Time                   `json:"program_created_at"`
	ProgramUpdatedAt time.Time                   `json:"program_updated_at"`
	CategoryTags     []ProgramDetailsCategoryTag `json:"category_tags"`
	Performers       []ProgramDetailsPerformer   `json:"performers"`
}

type programDetailsPerformerRaw struct {
	ID            int64   `json:"id"`
	FirstName     string  `json:"first_name"`
	LastName      string  `json:"last_name"`
	FirstNameKana string  `json:"first_name_kana"`
	LastNameKana  string  `json:"last_name_kana"`
	ImagePath     *string `json:"image_path"`
}

type ProgramsUsecase struct {
	q *db.Queries
}

func NewProgramsUsecase(q *db.Queries) *ProgramsUsecase {
	return &ProgramsUsecase{q: q}
}

func (u *ProgramsUsecase) GetProgramDetails(ctx context.Context, id int64) (ProgramDetails, error) {
	program, err := u.q.GetProgramByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ProgramDetails{}, ErrProgramNotFound
		}
		return ProgramDetails{}, err
	}

	categoryTagsJSON, err := normalizeJSONBytes(program.CategoryTags)
	if err != nil {
		return ProgramDetails{}, err
	}
	performersJSON, err := normalizeJSONBytes(program.Performers)
	if err != nil {
		return ProgramDetails{}, err
	}

	var categoryTags []ProgramDetailsCategoryTag
	if err := json.Unmarshal(categoryTagsJSON, &categoryTags); err != nil {
		return ProgramDetails{}, err
	}

	var performersRaw []programDetailsPerformerRaw
	if err := json.Unmarshal(performersJSON, &performersRaw); err != nil {
		return ProgramDetails{}, err
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

	resp := ProgramDetails{
		ProgramID:        program.ProgramID,
		Title:            program.Title,
		VideoURL:         buildVideoURL(program.VideoPath),
		ThumbnailUrl:    buildPublicFileURLPtr(nullStringPtr(program.ThumbnailPath)),
		Description:      nullStringPtr(program.Description),
		ProgramCreatedAt: program.ProgramCreatedAt,
		ProgramUpdatedAt: program.ProgramUpdatedAt,
		CategoryTags:     categoryTags,
		Performers:       performers,
	}
	return resp, nil
}

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
	println(base)
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
