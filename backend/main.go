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
	err := godotenv.Load(".env")
	// もし err がnilではないなら、"読み込み出来ませんでした"が出力されます。
	if err != nil {
		fmt.Printf("読み込み出来ませんでした: %v", err)
	} 

	ctx := context.Background()
	conn, err := dbconn.Open(ctx)
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	defer conn.Close()

	q := db.New(conn)
	r := router.NewRouter(q)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}
