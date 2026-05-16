package domain

type RegisterRequest struct {
	LastName   string
	FirstName  string
	MiddleName *string
	Email      string
	Password   string
}
