package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/GuilhAndrad/bookclub/config"
	"github.com/GuilhAndrad/bookclub/internal/handler"
	"github.com/GuilhAndrad/bookclub/internal/middleware"
	"github.com/GuilhAndrad/bookclub/internal/repository"
	"github.com/GuilhAndrad/bookclub/internal/service"
	"github.com/GuilhAndrad/bookclub/pkg/tokenblacklist"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
	"gorm.io/gorm"
)

// New constrói o roteador Gin com todas as dependências injetadas e rotas registradas.
// ctx deve ser o contexto raiz da aplicação — repassado às goroutines internas
// (blacklist e rate limiter) para garantir shutdown limpo quando cancelado.
func New(ctx context.Context, db *gorm.DB, cfg *config.Config) *gin.Engine {
	// gin.New em vez de gin.Default: middlewares registrados explicitamente
	// para controle total sobre formato de log e recovery.
	r := gin.New()
	r.Use(
		gin.Recovery(),      // captura panics e retorna 500 sem derrubar o servidor
		middleware.Logger(), // log estruturado por requisição via slog
		corsMiddleware(),
	)

	// Blacklist com limpeza a cada 15 minutos. Encerra quando ctx for cancelado.
	bl := tokenblacklist.New(ctx, 15*time.Minute)

	// Repositories
	userRepo := repository.NewUserRepository(db)
	bookRepo := repository.NewBookRepository(db)
	reviewRepo := repository.NewReviewRepository(db)
	commentRepo := repository.NewCommentRepository(db)

	// Services
	authSvc := service.NewAuthService(userRepo, cfg, bl)
	userSvc := service.NewUserService(userRepo)
	bookSvc := service.NewBookService(bookRepo)
	reviewSvc := service.NewReviewService(reviewRepo, commentRepo)

	// Middlewares
	authMiddleware := middleware.NewAuthMiddleware(authSvc)

	// Handlers
	authH := handler.NewAuthHandler(authSvc)
	userH := handler.NewUserHandler(userSvc)
	bookH := handler.NewBookHandler(bookSvc)
	reviewH := handler.NewReviewHandler(reviewSvc)

	// Health check — sem rate limit, usado por probes de infraestrutura.
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Auth pública — rate limit restritivo contra brute force.
	auth := r.Group("/auth")
	auth.Use(middleware.RateLimiter(ctx, rate.Limit(10), 20))
	{
		auth.POST("/register", authH.Register)
		auth.POST("/login", authH.Login)
	}

	// Rotas autenticadas — rate limit geral + JWT obrigatório.
	api := r.Group("/")
	api.Use(
		middleware.RateLimiter(ctx, rate.Limit(60), 120),
		authMiddleware.AuthRequired(),
	)
	{
		api.POST("/auth/logout", authH.Logout)
		api.POST("/auth/refresh", authH.RefreshToken)

		// Livros — ?page=1&limit=20
		api.GET("/books", bookH.ListBooks)
		api.GET("/books/current", bookH.GetCurrentBook)
		api.GET("/books/:id", bookH.GetBook)

		// Resenhas — ?page=1&limit=20
		api.GET("/books/:id/reviews", reviewH.ListReviews)
		api.POST("/books/:id/reviews", reviewH.CreateReview)
		api.PUT("/reviews/:id", reviewH.UpdateReview)
		api.DELETE("/reviews/:id", reviewH.DeleteReview)

		// Likes
		api.POST("/reviews/:id/like", reviewH.LikeReview)
		api.DELETE("/reviews/:id/like", reviewH.UnlikeReview)

		// Comentários — ?page=1&limit=20
		api.GET("/reviews/:id/comments", reviewH.ListComments)
		api.POST("/reviews/:id/comments", reviewH.CreateComment)

		// Perfil
		api.GET("/users/:id/profile", userH.GetProfile)
		api.GET("/users/:id/reviews", reviewH.ListUserReviews)
		api.PUT("/users/me/profile", userH.UpdateProfile)

		// Admin
		admin := api.Group("/admin")
		admin.Use(authMiddleware.AdminRequired())
		{
			admin.GET("/members/pending", userH.GetPendingMembers)
			admin.PUT("/members/:id/approve", userH.ApproveMember)
			admin.PUT("/members/:id/reject", userH.RejectMember)
			admin.POST("/books", bookH.CreateBook)
			admin.PUT("/books/:id", bookH.UpdateBook)
		}
	}

	slog.Info("rotas registradas",
		"public", 2,
		"authenticated", len(r.Routes())-3, // desconta /health e as 2 de auth
	)

	return r
}

// corsMiddleware adiciona os cabeçalhos CORS necessários para consumo pelo app Flutter.
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
