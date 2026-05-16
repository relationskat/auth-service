package domain

import "github.com/google/uuid"

type RegisterResponse struct {
	ID    string
	Email string
}

type User struct {
	ID         uuid.UUID
	LastName   string
	FirstName  string
	MiddleName *string
}
