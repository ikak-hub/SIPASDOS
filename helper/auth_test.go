package helper

import (
	"strings"
	"testing"
	"time"
)

var testSecret = strings.Repeat("s", 32)

func TestTokenRoundTrip(t *testing.T) {
	token, expires, err := GenerateToken(testSecret, time.Hour, 7, RoleAdmin)
	if err != nil {
		t.Fatalf("GenerateToken gagal: %v", err)
	}
	if !expires.After(time.Now()) {
		t.Error("token seharusnya kedaluwarsa di masa depan")
	}

	claims, err := ParseToken(testSecret, token)
	if err != nil {
		t.Fatalf("ParseToken gagal: %v", err)
	}
	if claims.UserID != 7 || claims.Role != RoleAdmin {
		t.Errorf("isi token salah: %+v", claims)
	}
}

func TestParseTokenRejectsInvalid(t *testing.T) {
	valid, _, _ := GenerateToken(testSecret, time.Hour, 1, RoleStudent)
	expired, _, _ := GenerateToken(testSecret, -time.Minute, 1, RoleStudent)

	cases := []struct {
		name, secret, token string
	}{
		{"secret berbeda", strings.Repeat("x", 32), valid},
		{"sudah kedaluwarsa", testSecret, expired},
		{"bukan token", testSecret, "bukan-jwt"},
		{"kosong", testSecret, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := ParseToken(tc.secret, tc.token); err == nil {
				t.Error("token seharusnya ditolak")
			}
		})
	}
}

func TestPasswordHash(t *testing.T) {
	hash, err := HashPassword("rahasia123")
	if err != nil {
		t.Fatalf("HashPassword gagal: %v", err)
	}
	if hash == "rahasia123" {
		t.Error("password tidak boleh disimpan apa adanya")
	}
	if !CheckPassword(hash, "rahasia123") {
		t.Error("password yang benar seharusnya lolos")
	}
	if CheckPassword(hash, "salah12345") {
		t.Error("password yang salah seharusnya ditolak")
	}
}
