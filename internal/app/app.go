// Package app wires the HTTP API together.
package app

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"libratrack/internal/handler"
	"libratrack/internal/middleware"
	"libratrack/internal/model"
	"libratrack/internal/repository"
)

func init() {
	// Release mode: no debug banner or route table in the logs.
	gin.SetMode(gin.ReleaseMode)
}

// NewRouter builds the router: it creates the repositories on top of db,
// the handlers on top of the repositories, and registers all routes.
func NewRouter(db *gorm.DB) http.Handler {
	categories := handler.NewCategoryHandler(repository.NewCategoryRepository(db))
	authors := handler.NewAuthorHandler(repository.NewAuthorRepository(db))
	books := handler.NewBookHandler(repository.NewBookRepository(db))
	copies := handler.NewCopyHandler(repository.NewCopyRepository(db))
	usersRepo := repository.NewUserRepository(db)
	authH := handler.NewAuthHandler(usersRepo)
	users := handler.NewUserHandler(usersRepo)
	loans := handler.NewLoanHandler(repository.NewLoanRepository(db))

	r := gin.New()
	// Order matters: Logger is the outermost so it sees the 500 after a panic;
	// Recovery comes before CORS; CORS sets headers before the handler runs.
	r.Use(middleware.Logger(), middleware.Recovery(), middleware.CORS())
	// A trailing slash is a different, unregistered route (JSON 404), not a 307 redirect.
	r.RedirectTrailingSlash = false
	r.NoRoute(handler.NotFound)
	r.GET("/health", handler.Health)
	r.GET("/openapi.yaml", handler.OpenAPI)
	r.GET("/swagger", handler.SwaggerUI)
	r.GET("/swagger/index.html", handler.SwaggerUI)

	api := r.Group("/api/v1")

	// Групи маршрутів. Кожна група успадковує middleware батьківської,
	// тому ПОРЯДОК вкладення визначає порядок виконання:
	//   public    — без токена;
	//   protected — AuthMiddleware (401, якщо токен відсутній/недійсний);
	//   staff     — protected + RequireRole(librarian, admin) (403 для інших);
	//   admin     — protected + RequireRole(admin).
	// Auth завжди ПЕРЕД RBAC: RequireRole читає роль, яку кладе AuthMiddleware.
	public := api.Group("")
	protected := api.Group("", middleware.AuthMiddleware())
	staff := protected.Group("", middleware.RequireRole(model.RoleLibrarian, model.RoleAdmin))
	admin := protected.Group("", middleware.RequireRole(model.RoleAdmin))

	authH.Register(public) // POST /auth/register, POST /auth/login
	categories.RegisterPublic(public)
	authors.RegisterPublic(public)
	books.RegisterPublic(public)
	copies.RegisterPublic(public)

	categories.RegisterStaff(staff)
	authors.RegisterStaff(staff)
	books.RegisterStaff(staff)
	copies.RegisterStaff(staff)
	loans.RegisterStaff(staff) // POST /loans, POST /loans/:id/return

	loans.RegisterProtected(protected) // GET /me/loans
	users.RegisterProtected(protected) // GET/PUT /me

	users.RegisterAdmin(admin) // GET /users, PUT /users/:id/role

	return r
}
