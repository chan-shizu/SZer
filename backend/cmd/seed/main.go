package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/chan-shizu/SZer/db"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/scrypt"
	"golang.org/x/text/unicode/norm"
)

const (
	scryptN     = 16384
	scryptR     = 16
	scryptP     = 1
	scryptKeyLn = 64
)

func hashBetterAuthPassword(password string) (string, error) {
	// better-auth uses password.normalize("NFKC")
	normalizedPassword := norm.NFKC.String(password)

	saltRaw := make([]byte, 16)
	if _, err := rand.Read(saltRaw); err != nil {
		return "", err
	}
	// NOTE: better-auth uses the hex string itself as the scrypt salt input (utf-8 bytes), not the raw 16 bytes.
	saltHex := hex.EncodeToString(saltRaw)

	key, err := scrypt.Key([]byte(normalizedPassword), []byte(saltHex), scryptN, scryptR, scryptP, scryptKeyLn)
	if err != nil {
		return "", err
	}
	keyHex := hex.EncodeToString(key)
	return saltHex + ":" + keyHex, nil
}

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

	// Seed users (better-auth tables)
	// NOTE: id is text primary key; for seed we use deterministic IDs.
	seedUsers := []struct {
		ID    string
		Name  string
		Email string
		Password string
	}{
		{ID: "seed-user-1", Name: "Seed User 1", Email: "seed1@example.com", Password: "password"},
		{ID: "seed-user-2", Name: "Seed User 2", Email: "seed2@example.com", Password: "password"},
	}
	for _, u := range seedUsers {
		if _, err := q.CreateAuthUser(ctx, db.CreateAuthUserParams{
			ID:            u.ID,
			Name:          u.Name,
			Email:         u.Email,
			EmailVerified: false,
			Image:         sql.NullString{Valid: false},
		}); err != nil {
			log.Fatalf("failed to create user %s: %v", u.ID, err)
		}

		passwordHash, err := hashBetterAuthPassword(u.Password)
		if err != nil {
			log.Fatalf("failed to hash password for user %s: %v", u.ID, err)
		}
		accountID := u.ID // better-auth uses accountId = userId for providerId="credential"
		accountRowID := "seed-account-" + u.ID
		if _, err := q.CreateCredentialAccount(ctx, db.CreateCredentialAccountParams{
			ID:        accountRowID,
			AccountId: accountID,
			UserId:    u.ID,
			Password:  sql.NullString{String: passwordHash, Valid: true},
		}); err != nil {
			log.Fatalf("failed to create credential account for user %s: %v", u.ID, err)
		}
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

		// keep for watch_histories seed later
		// program IDs start from 1 due to RESTART IDENTITY
		_ = program

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

		// Add a comment (user_id付き)
		var seedUserID string
		if i%2 == 0 {
			seedUserID = "seed-user-1"
		} else {
			seedUserID = "seed-user-2"
		}
		if _, err := q.CreateComment(ctx, db.CreateCommentParams{ProgramID: program.ID, UserID: sql.NullString{String: seedUserID, Valid: true}, Content: "Great video!"}); err != nil {
			log.Fatalf("failed to create comment: %v", err)
		}

		log.Printf("seed created video id=%d, title=%s", program.ID, vidParams.Title)
	}

	// Seed watch histories
	// Spec: view_count counts even incomplete watches, so we insert a few incomplete and completed entries.
	watchSeeds := []db.UpsertWatchHistoryParams{
		// program 1: two views (one incomplete, one completed)
		{UserID: "seed-user-1", ProgramID: 1, PositionSeconds: 120, IsCompleted: false},
		{UserID: "seed-user-2", ProgramID: 1, PositionSeconds: 600, IsCompleted: true},
		// program 2: one view (completed)
		{UserID: "seed-user-1", ProgramID: 2, PositionSeconds: 900, IsCompleted: true},
		// program 3: one view (incomplete)
		{UserID: "seed-user-2", ProgramID: 3, PositionSeconds: 42, IsCompleted: false},
	}
	for _, wh := range watchSeeds {
		if _, err := q.UpsertWatchHistory(ctx, wh); err != nil {
			log.Fatalf("failed to upsert watch history user_id=%s program_id=%d: %v", wh.UserID, wh.ProgramID, err)
		}
	}

	// Seed likes
	likeSeeds := []db.CreateLikeParams{
		{UserID: "seed-user-1", ProgramID: 1},
		{UserID: "seed-user-1", ProgramID: 2},
		{UserID: "seed-user-2", ProgramID: 1},
	}
	for _, l := range likeSeeds {
		if err := q.CreateLike(ctx, l); err != nil {
			log.Fatalf("failed to create like user_id=%s program_id=%d: %v", l.UserID, l.ProgramID, err)
		}
	}

	log.Println("seed completed")
}