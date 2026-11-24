package service

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	"sipasdos/app/repository"
	"sipasdos/helper"
)

func translateError(c *fiber.Ctx, err error, entity, generalMessage string) error {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return helper.Fail(c, fiber.StatusNotFound, entity+" tidak ditemukan")
	case errors.Is(err, repository.ErrDuplicate):
		return helper.Fail(c, fiber.StatusConflict, entity+" sudah ada")
	case errors.Is(err, repository.ErrInUse):
		return helper.Fail(c, fiber.StatusConflict, entity+" masih dipakai data lain")
	default:
		return helper.Fail(c, fiber.StatusInternalServerError, generalMessage)
	}
}
