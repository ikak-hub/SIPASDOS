package service

import (
	"testing"
	"time"

	"sipasdos/app/model"
)

func TestCountTotalPages(t *testing.T) {
	cases := []struct{ total, limit, want int }{
		{0, 10, 0},
		{1, 10, 1},
		{10, 10, 1},
		{11, 10, 2},
		{137, 20, 7},
	}
	for _, tc := range cases {
		if got := CountTotalPages(tc.total, tc.limit); got != tc.want {
			t.Errorf("total=%d limit=%d: harap %d, dapat %d", tc.total, tc.limit, tc.want, got)
		}
	}
}

func TestValidateCreatePeriod(t *testing.T) {
	t.Run("permintaan valid", func(t *testing.T) {
		period, errs := ValidateCreatePeriod(model.CreatePeriodRequest{
			Name: "Asdos Genap 2026", StartDate: "2026-12-01", EndDate: "2026-12-15",
		})
		if len(errs) != 0 {
			t.Fatalf("tidak seharusnya ada error: %v", errs)
		}
		if period.Status != model.PeriodDraft {
			t.Errorf("status awal harus draft, dapat %q", period.Status)
		}
	})

	cases := []struct {
		name      string
		req       model.CreatePeriodRequest
		wantField string
	}{
		{"nama kosong", model.CreatePeriodRequest{Name: "  ", StartDate: "2026-12-01", EndDate: "2026-12-15"}, "name"},
		{"format tanggal mulai salah", model.CreatePeriodRequest{Name: "A", StartDate: "01-12-2026", EndDate: "2026-12-15"}, "start_date"},
		{"tanggal akhir kosong", model.CreatePeriodRequest{Name: "A", StartDate: "2026-12-01"}, "end_date"},
		{"akhir sebelum awal", model.CreatePeriodRequest{Name: "A", StartDate: "2026-12-15", EndDate: "2026-12-01"}, "end_date"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, errs := ValidateCreatePeriod(tc.req)
			if _, ok := errs[tc.wantField]; !ok {
				t.Errorf("harap ada error pada field %q, dapat %v", tc.wantField, errs)
			}
		})
	}
}

func TestValidateReplacePeriod(t *testing.T) {
	current := model.RecruitmentPeriod{
		ID: 1, Name: "Lama", Status: model.PeriodDraft,
		StartDate: time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2026, 12, 15, 0, 0, 0, 0, time.UTC),
	}

	updated, errs := ValidateReplacePeriod(current, model.ReplacePeriodRequest{
		Name: "Baru", StartDate: "2026-12-02", EndDate: "2026-12-20", Status: model.PeriodOpen,
	})
	if len(errs) != 0 {
		t.Fatalf("tidak seharusnya ada error: %v", errs)
	}
	if updated.Name != "Baru" || updated.Status != model.PeriodOpen {
		t.Errorf("isi tidak terganti: %+v", updated)
	}

	_, errs = ValidateReplacePeriod(current, model.ReplacePeriodRequest{
		Name: "Baru", StartDate: "2026-12-02", EndDate: "2026-12-20", Status: model.PeriodClosed,
	})
	if _, ok := errs["status"]; !ok {
		t.Error("draft -> closed seharusnya ditolak")
	}
}

func TestApplyPatchPeriod(t *testing.T) {
	initial := model.RecruitmentPeriod{
		ID: 1, Name: "Asdos Ganjil", Status: model.PeriodDraft,
		StartDate: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC),
	}

	open := model.PeriodOpen
	result, errs := ApplyPatchPeriod(initial, model.PatchPeriodRequest{Status: &open})
	if len(errs) != 0 {
		t.Fatalf("tidak seharusnya ada error: %v", errs)
	}
	if result.Status != model.PeriodOpen {
		t.Error("status seharusnya berubah menjadi open")
	}
	if result.Name != "Asdos Ganjil" {
		t.Error("field yang tidak dikirim seharusnya tidak berubah")
	}

	closed := model.PeriodClosed
	if _, errs = ApplyPatchPeriod(initial, model.PatchPeriodRequest{Status: &closed}); errs["status"] == "" {
		t.Error("draft -> closed seharusnya ditolak")
	}

	early := "2026-07-01"
	if _, errs = ApplyPatchPeriod(initial, model.PatchPeriodRequest{EndDate: &early}); errs["end_date"] == "" {
		t.Error("tanggal akhir sebelum tanggal mulai seharusnya ditolak")
	}
}

func TestCanTransitionPeriod(t *testing.T) {
	cases := []struct {
		from, to string
		want     bool
	}{
		{model.PeriodDraft, model.PeriodOpen, true},
		{model.PeriodOpen, model.PeriodClosed, true},
		{model.PeriodDraft, model.PeriodClosed, false},
		{model.PeriodClosed, model.PeriodOpen, false},
		{model.PeriodOpen, model.PeriodOpen, true},
	}
	for _, tc := range cases {
		if got := CanTransitionPeriod(tc.from, tc.to); got != tc.want {
			t.Errorf("%s -> %s: harap %v, dapat %v", tc.from, tc.to, tc.want, got)
		}
	}
}

func TestIsEmptyPatchPeriod(t *testing.T) {
	if !IsEmptyPatchPeriod(model.PatchPeriodRequest{}) {
		t.Error("patch kosong harus terdeteksi")
	}
	name := "x"
	if IsEmptyPatchPeriod(model.PatchPeriodRequest{Name: &name}) {
		t.Error("patch berisi tidak boleh dianggap kosong")
	}
}
