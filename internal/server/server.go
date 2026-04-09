package server

import (
	"github.com/GuilhAndrad/bookclub/config"
	"github.com/GuilhAndrad/bookclub/internal/handler"
	"github.com/GuilhAndrad/bookclub/internal/middleware"
	"github.com/GuilhAndrad/bookclub/internal/repository"
	"github.com/GuilhAndrad/bookclub/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// New constrói o roteador Gin com todas as dependências injetadas e rotas registradas.
func New(db *gorm.DB, cfg *config.Config) *gin.Engine {
	r := gin.Default()
	r.Use(corsMiddleware())

	// Repositories
	userRepo      := repository.NewUserRepository(db)
	bookRepo      := repository.NewBookRepository(db)
	reviewRepo    := repository.NewReviewRepository(db)
	commentRepo   := repository.NewCommentRepository(db)
	bookshelfRepo := repository.NewBookshelfRepository(db)

	// Services
	authSvc      := service.NewAuthService(userRepo, cfg) // <- agora recebe a config
	userSvc      := service.NewUserService(userRepo)
	bookSvc      := service.NewBookService(bookRepo)
	reviewSvc    := service.NewReviewService(reviewRepo, commentRepo)
	bookshelfSvc := service.NewBookshelfService(bookshelfRepo)

	// Middlewares
	authMiddleware := middleware.NewAuthMiddleware(authSvc)

	// Handlers
	authH      := handler.NewAuthHandler(authSvc)
	userH      := handler.NewUserHandler(userSvc)
	bookH      := handler.NewBookHandler(bookSvc)
	reviewH    := handler.NewReviewHandler(reviewSvc)
	bookshelfH := handler.NewBookshelfHandler(bookshelfSvc)

	// Rotas públicas
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	auth := r.Group("/auth")
	{
		auth.POST("/register", authH.Register)
		auth.POST("/login", authH.Login)
	}

	// Rotas autenticadas
	api := r.Group("/")
	api.Use(authMiddleware.AuthRequired()) // <- método do AuthMiddleware
	{
		// Livros
		api.GET("/books", bookH.ListBooks)
		api.GET("/books/current", bookH.GetCurrentBook)
		api.GET("/books/:id", bookH.GetBook)

		// Resenhas
		api.GET("/books/:id/reviews", reviewH.ListReviews)
		api.POST("/books/:id/reviews", reviewH.CreateReview)
		api.PUT("/reviews/:id", reviewH.UpdateReview)
		api.DELETE("/reviews/:id", reviewH.DeleteReview)

		// Likes
		api.POST("/reviews/:id/like", reviewH.LikeReview)
		api.DELETE("/reviews/:id/like", reviewH.UnlikeReview)

		// Comentários
		api.GET("/reviews/:id/comments", reviewH.ListComments)
		api.POST("/reviews/:id/comments", reviewH.CreateComment)

		// Perfil
		api.GET("/users/:id/profile", userH.GetProfile)
		api.GET("/users/:id/reviews", reviewH.ListUserReviews)
		api.PUT("/users/me/profile", userH.UpdateProfile)

		// Estante virtual
		api.GET("/users/me/bookshelf", bookshelfH.GetShelf)
		api.PUT("/users/me/bookshelf/:bookId", bookshelfH.UpsertEntry)

		// Admin
		admin := api.Group("/admin")
		admin.Use(authMiddleware.AdminRequired()) // <- método do AuthMiddleware
		{
			admin.GET("/members/pending", userH.GetPendingMembers)
			admin.PUT("/members/:id/approve", userH.ApproveMember)
			admin.PUT("/members/:id/reject", userH.RejectMember)
			admin.POST("/books", bookH.CreateBook)
			admin.PUT("/books/:id", bookH.UpdateBook)
		}
	}

	return r
}

// corsMiddleware adiciona os cabeçalhos CORS necessários.
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}