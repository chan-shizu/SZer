package router

import (
	"github.com/chan-shizu/SZer/db"
	"github.com/chan-shizu/SZer/internal/handler"
	"github.com/chan-shizu/SZer/internal/middleware"
	"github.com/chan-shizu/SZer/internal/usecase"
	"github.com/gin-gonic/gin"
)

func NewRouter(q *db.Queries) *gin.Engine {
	router := gin.Default()
	programsUC := usecase.NewProgramsUsecase(q)
	usersUC := usecase.NewUsersUsecase(q)
	h := handler.New(programsUC, usersUC)
	commentsHandler := handler.NewCommentsHandler(q)

	// 認証不要のエンドポイント
	router.GET("/top", h.Top)
	router.GET("/top/liked", h.TopLiked)
	router.GET("/top/viewed", h.TopViewed)
	router.GET("/programs/:id", middleware.OptionalAuth(), h.ProgramDetails)
	router.GET("/programs", h.ListPrograms)
	
	// コメントAPI（未ログインOK）
	router.GET("/programs/:id/comments", middleware.OptionalAuth(), commentsHandler.ListComments)
	router.POST("/programs/:id/comments", middleware.OptionalAuth(), commentsHandler.PostComment)
	
	// マイページ系APIのみ認証必須
	authenticated := router.Group("/")
	authenticated.Use(middleware.RequireAuth())
	authenticated.POST("watch-histories", h.UpsertWatchHistory)
	authenticated.POST("programs/:id/like", h.LikeProgram)
	authenticated.DELETE("programs/:id/like", h.UnlikeProgram)
	authenticated.GET("me/watching-programs", h.ListWatchingPrograms)
	authenticated.GET("me/liked-programs", h.ListLikedPrograms)
	authenticated.GET("me/points", h.GetPoints)
	authenticated.POST("me/points/add", h.AddPoints)

	return router
}
