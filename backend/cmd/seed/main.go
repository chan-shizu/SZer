package main

import (
	"context"
	"database/sql"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/chan-shizu/SZer/db"
	_ "github.com/lib/pq"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:password@localhost:5432/postgres?sslmode=disable"
	}

	// lib/pq defaults to sslmode=require when not provided.
	// Our local/docker Postgres typically does not have SSL enabled, so force-disable
	// SSL unless the user explicitly set sslmode.
	if !strings.Contains(dsn, "sslmode=") {
		if u, err := url.Parse(dsn); err == nil {
			q := u.Query()
			if q.Get("sslmode") == "" {
				q.Set("sslmode", "disable")
				u.RawQuery = q.Encode()
				dsn = u.String()
			}
		}
	}

	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := conn.PingContext(ctx); err != nil {
		log.Fatalf("failed to ping db: %v", err)
	}

	q := db.New(conn)

	// 元のすべてのデータをクリア
	if err := q.ClearAllData(ctx); err != nil {
		log.Fatalf("failed to clear data: %v", err)
	}

	// Seed category tags
	tagNames := []string{"音楽", "お笑い", "グルメ", "その他"}
	var tags []db.CategoryTag
	for _, n := range tagNames {
		t, err := q.CreateCategoryTag(ctx, n)
		if err != nil {
			log.Fatalf("failed to create tag %s: %v", n, err)
		}
		tags = append(tags, t)
	}

	// Seed performers
	performerParams := []db.CreatePerformerParams{
		{FirstName: "光一郎", LastName: "靜谷", FirstNameKana: "コウイチロウ", LastNameKana: "シズヤ", ImagePath: sql.NullString{String: "performer/shizuya.jpg", Valid: true}},
	}
	var performers []db.Performer
	for _, p := range performerParams {
		perf, err := q.CreatePerformer(ctx, p)
		if err != nil {
			log.Fatalf("failed to create performer: %v", err)
		}
		performers = append(performers, perf)
	}

	// Seed sample videos
	programParamsList := []db.CreateProgramParams{
		{
			Title:         "所沢の車全部壊してみた！",
			VideoPath:     "ReInventAI.mp4",
			ThumbnailPath: sql.NullString{String: "thumbnail/sample.jpg", Valid: true},
			Description:   sql.NullString{String: "This is a sample seeded video.", Valid: true},
		},
		{
			Title:         "美味しいラーメンの作り方",
			VideoPath:     "ReInventAI.mp4",
			ThumbnailPath: sql.NullString{String: "thumbnail/sample.jpg", Valid: true},
			Description:   sql.NullString{String: "本格的なラーメンのレシピを紹介します。", Valid: true},
		},
		{
			Title:         "ギター弾き語りライブ",
			VideoPath:     "ReInventAI.mp4",
			ThumbnailPath: sql.NullString{String: "thumbnail/sample.jpg", Valid: true},
			Description:   sql.NullString{String: "週末のライブ映像です。", Valid: true},
		},
		{
			Title:         "お笑いライブ2024",
			VideoPath:     "ReInventAI.mp4",
			ThumbnailPath: sql.NullString{String: "thumbnail/sample.jpg", Valid: true},
			Description:   sql.NullString{String: "最新のお笑いライブ映像をお届けします。", Valid: true},
		},
	}

	for i, vidParams := range programParamsList {
		program, err := q.CreateProgram(ctx, vidParams)
		if err != nil {
			log.Fatalf("failed to create video: %v", err)
		}

		// Link tags to video
		limit := i + 1
		if limit > len(tags) {
			limit = len(tags)
		}
		for _, t := range tags[:limit] {
			if err := q.CreateProgramCategoryTag(ctx, db.CreateProgramCategoryTagParams{ProgramID: program.ID, TagID: t.ID}); err != nil {
				log.Fatalf("failed to link tag %d to video %d: %v", t.ID, program.ID, err)
			}
		}

		// Link performers to video
		for _, p := range performers {
			if err := q.CreateProgramPerformer(ctx, db.CreateProgramPerformerParams{ProgramID: program.ID, PerformerID: p.ID}); err != nil {
				log.Fatalf("failed to link performer %d to video %d: %v", p.ID, program.ID, err)
			}
		}

		// Add a comment
		if _, err := q.CreateComment(ctx, db.CreateCommentParams{ProgramID: program.ID, Content: "Great video!"}); err != nil {
			log.Fatalf("failed to create comment: %v", err)
		}

		log.Printf("seed created video id=%d, title=%s", program.ID, vidParams.Title)
	}

	log.Println("seed completed")
}