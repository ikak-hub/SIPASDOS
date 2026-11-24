package model

import "time"

const (
	RoleAdmin    = "admin"
	RoleLecturer = "dosen"
	RoleStudent  = "mahasiswa"
)

type User struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}
