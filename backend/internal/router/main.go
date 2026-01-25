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

	// Require better-auth session for backend APIs.
	router.Use(middleware.RequireAuth())

	router.GET("/top", h.Top)
	router.GET("/top/liked", h.TopLiked)
	router.GET("/top/viewed", h.TopViewed)
	router.GET("/programs/:id", h.ProgramDetails)
	router.POST("/programs/:id/like", h.LikeProgram)
	router.DELETE("/programs/:id/like", h.UnlikeProgram)
	router.GET("/programs", h.ListPrograms)
	router.GET("/me/watching-programs", h.ListWatchingPrograms)
	router.GET("/me/liked-programs", h.ListLikedPrograms)
	router.GET("/me/points", h.GetPoints)
	router.POST("/me/points/add", h.AddPoints)
	router.POST("/watch-histories", h.UpsertWatchHistory)

	return router
}
