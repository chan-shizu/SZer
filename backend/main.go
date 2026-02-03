package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/chan-shizu/SZer/db"
	"github.com/chan-shizu/SZer/internal/dbconn"
	"github.com/chan-shizu/SZer/internal/router"
	"github.com/joho/godotenv"
)

func main() {
	// .envファイルを読み込む（環境変数として使えるようにする）
	if err := godotenv.Load(".env"); err != nil {
		fmt.Printf(".env読み込み失敗: %v\n", err)
	}

	ctx := context.Background()
	conn, err := dbconn.Open(ctx)
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	defer conn.Close()

	q := db.New(conn)
	r := router.NewRouter(conn, q)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}
