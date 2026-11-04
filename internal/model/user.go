// Package model berisi struct yang memetakan tabel database (db/schema.sql).
// FILE BERSAMA (frozen): ubah hanya lewat PR bersama bila schema berubah.
package model

import "time"

type Role string

const (
	RoleAdmin   Role = "admin"
	RolePekerja Role = "pekerja"
)

// User memetakan tabel `users`.
type User struct {
	ID           int64
	Username     string
	PasswordHash string
	Role         Role
	Status       string // aktif | nonaktif
	DibuatPada   time.Time
	DiubahPada   time.Time
}

// SessionUser adalah identitas yang disimpan di session dan dibawa di context request.
type SessionUser struct {
	UserID    int64
	Username  string
	Nama      string
	Role      Role
	PekerjaID int64 // 0 untuk admin
}

func (u SessionUser) IsAdmin() bool   { return u.Role == RoleAdmin }
func (u SessionUser) IsPekerja() bool { return u.Role == RolePekerja }
