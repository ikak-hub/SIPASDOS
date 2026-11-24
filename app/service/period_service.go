package service

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	"sipasdos/app/model"
	"sipasdos/app/repository"
	"sipasdos/helper"
)

const periodEntity = "periode rekrutmen"

type PeriodService struct {
	repo repository.PeriodRepository
}

func NewPeriodService(repo repository.PeriodRepository) *PeriodService {
	return &PeriodService{repo: repo}
}

func (s *PeriodService) List(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	q := helper.ParseListQuery(c, PeriodSortable)

	periods, total, err := s.repo.FindAll(ctx, q)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError,
			"gagal mengambil data periode rekrutmen")
	}

	return helper.SuccessList(c, "daftar periode rekrutmen berhasil diambil", periods, &model.Meta{
		Page:       q.Page,
		Limit:      q.Limit,
		Total:      total,
		TotalPages: CountTotalPages(total, q.Limit),
	})
}

func (s *PeriodService) Get(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	period, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateError(c, err, periodEntity, "gagal mengambil data periode rekrutmen")
	}
	return helper.Success(c, fiber.StatusOK, "periode rekrutmen ditemukan", period)
}

func (s *PeriodService) Create(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	var req model.CreatePeriodRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	// Business rulesnya dipanggil, bukan ditulis ulang di sini.
	period, errs := ValidateCreatePeriod(req)
	if len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}

	created, err := s.repo.Create(ctx, period)
	if err != nil {
		return translateError(c, err, periodEntity, "gagal menyimpan periode rekrutmen")
	}
	return helper.Created(c, "periode rekrutmen berhasil dibuat", created,
		"/api/v1/periods/"+strconv.Itoa(created.ID))
}

func (s *PeriodService) Replace(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	var req model.ReplacePeriodRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	current, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateError(c, err, periodEntity, "gagal mengambil data periode rekrutmen")
	}

	updated, errs := ValidateReplacePeriod(current, req)
	if len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}

	result, err := s.repo.Update(ctx, updated)
	if err != nil {
		return translateError(c, err, periodEntity, "gagal memperbarui periode rekrutmen")
	}
	return helper.Success(c, fiber.StatusOK, "periode rekrutmen berhasil diganti seluruhnya", result)
}

func (s *PeriodService) Patch(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	var req model.PatchPeriodRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}
	if IsEmptyPatchPeriod(req) {
		return helper.Fail(c, fiber.StatusBadRequest, "tidak ada field yang diubah")
	}

	current, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateError(c, err, periodEntity, "gagal mengambil data periode rekrutmen")
	}

	updated, errs := ApplyPatchPeriod(current, req)
	if len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}

	result, err := s.repo.Update(ctx, updated)
	if err != nil {
		return translateError(c, err, periodEntity, "gagal memperbarui periode rekrutmen")
	}
	return helper.Success(c, fiber.StatusOK, "periode rekrutmen berhasil diperbarui sebagian", result)
}

func (s *PeriodService) Delete(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return translateError(c, err, periodEntity, "gagal menghapus periode rekrutmen")
	}
	return helper.NoContent(c)
}
