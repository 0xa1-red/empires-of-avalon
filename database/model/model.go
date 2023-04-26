package model

import "github.com/google/uuid"

type UserRequest struct {
	Username string
	Email    string
	Password string
}

type User struct {
	ID       uuid.UUID
	Username string
	Email    string
	Hash     string
}
