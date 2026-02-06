package helper

import "github.com/google/uuid"

type User struct {
	ID      uuid.UUID
	Login   string
	Name    string
	Surname string
}
