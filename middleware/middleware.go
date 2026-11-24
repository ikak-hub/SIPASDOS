package middleware

import (
	"log/slog"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"

	"sipasdos/helper"
)

func Register(app *fiber.App, logger *slog.Logger) {
	app.Use(requestid.New())      
	app.Use(recover.New())         
	app.Use(helmet.New())         
	app.Use(cors.New())            
	app.Use(RequestLogger(logger)) 
}

func RequestLogger(logger *slog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next() // serahkan ke middleware/handler berikutnya

		requestID, _ := c.Locals("requestid").(string)
		logger.Info("http_request",
			slog.String("request_id", requestID),
			slog.String("method", c.Method()),
			slog.String("path", c.Path()),
			slog.Int("status", c.Response().StatusCode()),
			slog.Duration("duration", time.Since(start)),
			slog.String("ip", c.IP()),
		)
		return err
	}
}

var methodsWithBody = map[string]bool{
	fiber.MethodPost:  true,
	fiber.MethodPut:   true,
	fiber.MethodPatch: true,
}

func RequireJSON(c *fiber.Ctx) error {
	if methodsWithBody[c.Method()] {
		ct := c.Get("Content-Type")
		if !strings.HasPrefix(ct, fiber.MIMEApplicationJSON) {
			return helper.Fail(c, fiber.StatusUnsupportedMediaType,
				"Content-Type harus application/json")
		}
	}
	return c.Next()
}

func JWT(secret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		scheme, token, found := strings.Cut(c.Get(fiber.HeaderAuthorization), " ")
		token = strings.TrimSpace(token)
		if !found || !strings.EqualFold(scheme, "Bearer") || token == "" {
			c.Set(fiber.HeaderWWWAuthenticate, "Bearer")
			return helper.Fail(c, fiber.StatusUnauthorized, "token autentikasi diperlukan")
		}

		claims, err := helper.ParseToken(secret, token)
		if err != nil {
			c.Set(fiber.HeaderWWWAuthenticate, "Bearer")
			return helper.Fail(c, fiber.StatusUnauthorized, "token tidak valid atau kedaluwarsa")
		}

		c.Locals(helper.LocalUserID, claims.UserID)
		c.Locals(helper.LocalRole, claims.Role)
		return c.Next()
	}
}

func RequireRole(roles ...string) fiber.Handler {
	allowed := make(map[string]bool, len(roles))
	for _, role := range roles {
		allowed[role] = true
	}

	return func(c *fiber.Ctx) error {
		_, role, ok := helper.CurrentUser(c)
		if !ok {
			return helper.Fail(c, fiber.StatusUnauthorized, "token autentikasi diperlukan")
		}
		if !allowed[role] {
			return helper.Fail(c, fiber.StatusForbidden, "akses ditolak untuk role ini")
		}
		return c.Next()
	}
}
