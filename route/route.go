package route

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"sipasdos/app/service"
	"sipasdos/helper"
	"sipasdos/middleware"
)

type Services struct {
	Auth   *service.AuthService
	Period *service.PeriodService
	// Vacancy     *service.VacancyService
	// Application *service.ApplicationService
}

func Register(app *fiber.App, pool *pgxpool.Pool, s Services, authRequired fiber.Handler) {
	api := app.Group("/api/v1")
	adminOnly := middleware.RequireRole(helper.RoleAdmin)

	api.Get("/health", healthCheck(pool))

	auth := api.Group("/auth", middleware.RequireJSON)
	auth.Post("/register", s.Auth.Register)
	auth.Post("/login", s.Auth.Login)

	api.Get("/auth/me", authRequired, s.Auth.Me)

	periods := api.Group("/periods", authRequired, middleware.RequireJSON)
	periods.Get("/", s.Period.List)
	periods.Get("/:id", s.Period.Get)
	periods.Post("/", adminOnly, s.Period.Create)
	periods.Put("/:id", adminOnly, s.Period.Replace)
	periods.Patch("/:id", adminOnly, s.Period.Patch)
	periods.Delete("/:id", adminOnly, s.Period.Delete)
}

func healthCheck(pool *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
		defer cancel()

		if err := pool.Ping(ctx); err != nil {
			return helper.Fail(c, fiber.StatusServiceUnavailable,
				"database tidak dapat dihubungi")
		}
		return helper.Success(c, fiber.StatusOK, "server dan database berjalan", nil)
	}
}
