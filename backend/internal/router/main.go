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

	// 認証不要のエンドポイント
	router.GET("/top", h.Top)
	router.GET("/top/liked", h.TopLiked)
	router.GET("/top/viewed", h.TopViewed)
	router.GET("/programs/:id", h.ProgramDetails)
	router.POST("/programs/:id/like", h.LikeProgram)
	router.DELETE("/programs/:id/like", h.UnlikeProgram)
	router.GET("/programs", h.ListPrograms)
	router.POST("/watch-histories", h.UpsertWatchHistory)

	// マイページ系APIのみ認証必須
	me := router.Group("/me")
	me.Use(middleware.RequireAuth())
	me.GET("/watching-programs", h.ListWatchingPrograms)
	me.GET("/liked-programs", h.ListLikedPrograms)
	me.GET("/points", h.GetPoints)
	me.POST("/points/add", h.AddPoints)

	return router
}
