package router

import (
	"database/sql"
	"log"

	"github.com/chan-shizu/SZer/db"
	cfutil "github.com/chan-shizu/SZer/internal/cloudfront"
	"github.com/chan-shizu/SZer/internal/handler"
	"github.com/chan-shizu/SZer/internal/middleware"
	"github.com/chan-shizu/SZer/internal/usecase"
	"github.com/gin-gonic/gin"
)

func NewRouter(conn *sql.DB, q *db.Queries) *gin.Engine {
	router := gin.Default()

	signer, err := cfutil.NewVideoURLSigner()
	if err != nil {
		log.Fatalf("CloudFront signer初期化失敗: %v", err)
	}

	programsUC := usecase.NewProgramsUsecase(q, signer)
	paypayUC := usecase.NewPayPayUsecase(conn, q)

	requestsUC := usecase.NewRequestsUsecase(q)

	programsHandler := handler.NewProgramsHandler(programsUC)
	paypayHandler := handler.NewPayPayHandler(paypayUC)
	commentsHandler := handler.NewCommentsHandler(q)
	paypayWebhookHandler := handler.NewPayPayWebhookHandler(conn, q)
	requestsHandler := handler.NewRequestsHandler(requestsUC)

	
	// 認証不要のエンドポイント
	router.GET("/top", programsHandler.Top)
	router.GET("/top/liked", programsHandler.TopLiked)
	router.GET("/top/viewed", programsHandler.TopViewed)
	router.GET("/programs/:id", middleware.OptionalAuth(), programsHandler.ProgramDetails)
	router.GET("/programs", programsHandler.ListPrograms)

	// PayPay Webhook（認証不要）
	router.POST("/paypay/webhook", paypayWebhookHandler.Handle)

	// コメントAPI（未ログインOK）
	router.GET("/programs/:id/comments", middleware.OptionalAuth(), commentsHandler.ListComments)
	router.POST("/programs/:id/comments", middleware.OptionalAuth(), commentsHandler.PostComment)

	// リクエストAPI（未ログインOK）
	router.POST("/requests", middleware.OptionalAuth(), requestsHandler.CreateRequest)

	// マイページ系APIのみ認証必須
	authenticated := router.Group("/")
	authenticated.Use(middleware.RequireAuth())
	authenticated.POST("watch-histories", programsHandler.UpsertWatchHistory)
	authenticated.POST("programs/:id/like", programsHandler.LikeProgram)
	authenticated.DELETE("programs/:id/like", programsHandler.UnlikeProgram)
	authenticated.GET("me/watching-programs", programsHandler.ListWatchingPrograms)
	authenticated.GET("me/liked-programs", programsHandler.ListLikedPrograms)
	authenticated.GET("me/purchased-programs", programsHandler.ListPurchasedPrograms)
	authenticated.POST("/me/paypay/checkout", paypayHandler.PayPayCheckout)
	authenticated.GET("/me/paypay/payments/:merchantPaymentId", paypayHandler.PayPayGetPayment)

	return router
}
