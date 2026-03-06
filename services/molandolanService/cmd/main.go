package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/JokerTrickster/joker_backend/shared"
	"github.com/JokerTrickster/joker_backend/shared/db/mysql"
	"github.com/JokerTrickster/joker_backend/shared/logger"
	sharedMiddleware "github.com/JokerTrickster/joker_backend/shared/middleware"

	authHandler "github.com/JokerTrickster/joker_backend/services/molandolanService/features/auth/handler"
	authRepo "github.com/JokerTrickster/joker_backend/services/molandolanService/features/auth/repository"
	authUseCase "github.com/JokerTrickster/joker_backend/services/molandolanService/features/auth/usecase"

	newsHandler "github.com/JokerTrickster/joker_backend/services/molandolanService/features/news/handler"
	newsRepo "github.com/JokerTrickster/joker_backend/services/molandolanService/features/news/repository"
	newsUseCase "github.com/JokerTrickster/joker_backend/services/molandolanService/features/news/usecase"

	productHandler "github.com/JokerTrickster/joker_backend/services/molandolanService/features/product/handler"
	productRepo "github.com/JokerTrickster/joker_backend/services/molandolanService/features/product/repository"
	productUseCase "github.com/JokerTrickster/joker_backend/services/molandolanService/features/product/usecase"

	rankingHandler "github.com/JokerTrickster/joker_backend/services/molandolanService/features/ranking/handler"
	rankingRepo "github.com/JokerTrickster/joker_backend/services/molandolanService/features/ranking/repository"
	rankingUseCase "github.com/JokerTrickster/joker_backend/services/molandolanService/features/ranking/usecase"

	galleryHandler "github.com/JokerTrickster/joker_backend/services/molandolanService/features/gallery/handler"
	galleryRepo "github.com/JokerTrickster/joker_backend/services/molandolanService/features/gallery/repository"
	galleryUseCase "github.com/JokerTrickster/joker_backend/services/molandolanService/features/gallery/usecase"

	uploadHandler "github.com/JokerTrickster/joker_backend/services/molandolanService/features/upload/handler"
	uploadUseCase "github.com/JokerTrickster/joker_backend/services/molandolanService/features/upload/usecase"

	molandolanMiddleware "github.com/JokerTrickster/joker_backend/services/molandolanService/middleware"

	"go.uber.org/zap"
)

func main() {
	e, err := shared.Init(&shared.InitConfig{
		LogLevel:    os.Getenv("LOG_LEVEL"),
		Environment: os.Getenv("ENV"),
	})
	if err != nil {
		log.Fatal("Failed to initialize:", err)
	}
	defer shared.Cleanup()

	logger.Info("Starting Molandolan Service",
		zap.String("environment", shared.AppConfig.Env),
	)

	db := mysql.GormMysqlDB
	timeoutStr := os.Getenv("REQUEST_TIMEOUT")
	timeout := 10 * time.Second
	if timeoutStr != "" {
		if d, err := time.ParseDuration(timeoutStr); err == nil {
			timeout = d
		}
	}

	// Auth
	authRepository := authRepo.NewAuthRepository(db)
	oauthUC := authUseCase.NewOAuthUseCase(authRepository, timeout)
	meUC := authUseCase.NewMeUseCase(authRepository, timeout)
	updateMeUC := authUseCase.NewUpdateMeUseCase(authRepository, timeout)
	oauthH := authHandler.NewOAuthHandler(oauthUC)
	logoutH := authHandler.NewLogoutHandler()
	meH := authHandler.NewMeHandler(meUC)
	updateMeH := authHandler.NewUpdateMeHandler(updateMeUC)

	// News
	newsRepository := newsRepo.NewNewsRepository(db)
	newsListUC := newsUseCase.NewListUseCase(newsRepository, timeout)
	newsDetailUC := newsUseCase.NewDetailUseCase(newsRepository, timeout)
	newsCreateUC := newsUseCase.NewCreateUseCase(newsRepository, timeout)
	newsUpdateUC := newsUseCase.NewUpdateUseCase(newsRepository, timeout)
	newsDeleteUC := newsUseCase.NewDeleteUseCase(newsRepository, timeout)
	newsListH := newsHandler.NewListHandler(newsListUC)
	newsDetailH := newsHandler.NewDetailHandler(newsDetailUC)
	newsCreateH := newsHandler.NewCreateHandler(newsCreateUC)
	newsUpdateH := newsHandler.NewUpdateHandler(newsUpdateUC)
	newsDeleteH := newsHandler.NewDeleteHandler(newsDeleteUC)

	// Product
	productRepository := productRepo.NewProductRepository(db)
	productListUC := productUseCase.NewListUseCase(productRepository, timeout)
	productDetailUC := productUseCase.NewDetailUseCase(productRepository, timeout)
	productCreateUC := productUseCase.NewCreateUseCase(productRepository, timeout)
	productUpdateUC := productUseCase.NewUpdateUseCase(productRepository, timeout)
	productDeleteUC := productUseCase.NewDeleteUseCase(productRepository, timeout)
	productListH := productHandler.NewListHandler(productListUC)
	productDetailH := productHandler.NewDetailHandler(productDetailUC)
	productCreateH := productHandler.NewCreateHandler(productCreateUC)
	productUpdateH := productHandler.NewUpdateHandler(productUpdateUC)
	productDeleteH := productHandler.NewDeleteHandler(productDeleteUC)

	// Ranking
	rankingRepository := rankingRepo.NewRankingRepository(db)
	rankingListUC := rankingUseCase.NewListUseCase(rankingRepository, timeout)
	rankingMeUC := rankingUseCase.NewMeUseCase(rankingRepository, timeout)
	rankingSubmitUC := rankingUseCase.NewSubmitUseCase(rankingRepository, timeout)
	rankingDeleteUC := rankingUseCase.NewDeleteUseCase(rankingRepository, timeout)
	rankingListH := rankingHandler.NewListHandler(rankingListUC)
	rankingMeH := rankingHandler.NewMeHandler(rankingMeUC)
	rankingSubmitH := rankingHandler.NewSubmitHandler(rankingSubmitUC)
	rankingDeleteH := rankingHandler.NewDeleteHandler(rankingDeleteUC)

	// Gallery
	galleryRepository := galleryRepo.NewGalleryRepository(db)
	galleryListUC := galleryUseCase.NewListUseCase(galleryRepository, timeout)
	galleryDetailUC := galleryUseCase.NewDetailUseCase(galleryRepository, timeout)
	galleryCreateUC := galleryUseCase.NewCreateUseCase(galleryRepository, timeout)
	galleryDeleteUC := galleryUseCase.NewDeleteUseCase(galleryRepository, timeout)
	galleryLikeUC := galleryUseCase.NewLikeUseCase(galleryRepository, timeout)
	galleryCommentListUC := galleryUseCase.NewCommentListUseCase(galleryRepository, timeout)
	galleryCommentCreateUC := galleryUseCase.NewCommentCreateUseCase(galleryRepository, timeout)
	galleryCommentDeleteUC := galleryUseCase.NewCommentDeleteUseCase(galleryRepository, timeout)
	galleryListH := galleryHandler.NewListHandler(galleryListUC)
	galleryDetailH := galleryHandler.NewDetailHandler(galleryDetailUC)
	galleryCreateH := galleryHandler.NewCreateHandler(galleryCreateUC)
	galleryDeleteH := galleryHandler.NewDeleteHandler(galleryDeleteUC, galleryRepository)
	galleryLikeH := galleryHandler.NewLikeHandler(galleryLikeUC)
	galleryCommentListH := galleryHandler.NewCommentListHandler(galleryCommentListUC)
	galleryCommentCreateH := galleryHandler.NewCommentCreateHandler(galleryCommentCreateUC)
	galleryCommentDeleteH := galleryHandler.NewCommentDeleteHandler(galleryCommentDeleteUC, galleryRepository)

	// Upload
	uploadUC := uploadUseCase.NewUploadUseCase(timeout)
	uploadH := uploadHandler.NewUploadHandler(uploadUC)

	// Public routes (OAuth)
	e.GET("/api/auth/:provider", oauthH.Redirect)
	e.GET("/api/auth/:provider/callback", oauthH.Callback)
	e.GET("/api/news", newsListH.List)
	e.GET("/api/news/:id", newsDetailH.Detail)
	e.GET("/api/products", productListH.List)
	e.GET("/api/products/:id", productDetailH.Detail)
	e.GET("/api/rankings/:gameType", rankingListH.List)
	e.GET("/api/gallery", galleryListH.List, molandolanMiddleware.OptionalAuth())
	e.GET("/api/gallery/:id", galleryDetailH.Detail, molandolanMiddleware.OptionalAuth())
	e.GET("/api/gallery/:id/comments", galleryCommentListH.List)

	// User routes (JWT required)
	userGroup := e.Group("", sharedMiddleware.JWTAuth())
	userGroup.POST("/api/auth/logout", logoutH.Logout)
	userGroup.GET("/api/auth/me", meH.Me)
	userGroup.PUT("/api/auth/me", updateMeH.UpdateMe)
	userGroup.GET("/api/rankings/:gameType/me", rankingMeH.MyRanking)
	userGroup.POST("/api/rankings/:gameType", rankingSubmitH.Submit)
	userGroup.POST("/api/gallery", galleryCreateH.Create)
	userGroup.DELETE("/api/gallery/:id", galleryDeleteH.Delete)
	userGroup.POST("/api/gallery/:id/like", galleryLikeH.Like)
	userGroup.POST("/api/gallery/:id/comments", galleryCommentCreateH.Create)
	userGroup.DELETE("/api/gallery/:id/comments/:commentId", galleryCommentDeleteH.Delete)

	// Admin routes (JWT + Admin required)
	adminGroup := e.Group("", sharedMiddleware.JWTAuth(), molandolanMiddleware.AdminAuth(db))
	adminGroup.POST("/api/news", newsCreateH.Create)
	adminGroup.PUT("/api/news/:id", newsUpdateH.Update)
	adminGroup.DELETE("/api/news/:id", newsDeleteH.Delete)
	adminGroup.POST("/api/products", productCreateH.Create)
	adminGroup.PUT("/api/products/:id", productUpdateH.Update)
	adminGroup.DELETE("/api/products/:id", productDeleteH.Delete)
	adminGroup.DELETE("/api/rankings/:gameType/:id", rankingDeleteH.Delete)
	adminGroup.POST("/api/upload", uploadH.Upload)

	port := os.Getenv("PORT")
	if port == "" {
		port = "18083"
	}

	logger.Info("Server starting", zap.String("port", port))

	go func() {
		if err := e.Start(":" + port); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}
	logger.Info("Server exited gracefully")
}
