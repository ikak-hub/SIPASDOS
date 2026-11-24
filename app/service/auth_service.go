package service

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"

	"sipasdos/app/model"
	"sipasdos/app/repository"
	"sipasdos/helper"
)

const accountEntity = "akun"

type AuthService struct {
	repo   repository.UserRepository
	secret string
	ttl    time.Duration

	// dummyHash dipakai saat email tidak ditemukan agar waktu respons
	// login tidak membedakan email yang terdaftar dari yang tidak.
	dummyHash string
}

func NewAuthService(repo repository.UserRepository, secret string, ttl time.Duration) *AuthService {
	dummy, _ := helper.HashPassword("password-tidak-dipakai")
	return &AuthService{repo: repo, secret: secret, ttl: ttl, dummyHash: dummy}
}

func (s *AuthService) Register(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	var req model.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	user, errs := ValidateRegister(req)
	if len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}

	hash, err := helper.HashPassword(req.Password)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal memproses password")
	}
	user.PasswordHash = hash

	created, err := s.repo.Create(ctx, user)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return helper.Fail(c, fiber.StatusConflict, "email sudah terdaftar")
		}
		return translateError(c, err, accountEntity, "gagal mendaftarkan akun")
	}
	return helper.Success(c, fiber.StatusCreated, "akun berhasil dibuat", created)
}

func (s *AuthService) Login(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	var req model.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}
	if errs := ValidateLogin(req); len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}

	const invalidCredentials = "email atau password salah"

	user, err := s.repo.FindByEmail(ctx, NormalizeEmail(req.Email))
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			helper.CheckPassword(s.dummyHash, req.Password)
			return helper.Fail(c, fiber.StatusUnauthorized, invalidCredentials)
		}
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal memproses login")
	}
	if !helper.CheckPassword(user.PasswordHash, req.Password) {
		return helper.Fail(c, fiber.StatusUnauthorized, invalidCredentials)
	}

	token, expires, err := helper.GenerateToken(s.secret, s.ttl, user.ID, user.Role)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal membuat token")
	}

	return helper.Success(c, fiber.StatusOK, "login berhasil", model.AuthResponse{
		Token:     token,
		TokenType: "Bearer",
		ExpiresAt: expires,
		User:      user,
	})
}

func (s *AuthService) Me(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, _, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "token autentikasi diperlukan")
	}

	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateError(c, err, accountEntity, "gagal mengambil data akun")
	}
	return helper.Success(c, fiber.StatusOK, "data akun", user)
}
