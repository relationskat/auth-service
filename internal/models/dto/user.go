package dto

type User struct {
	ID           string
	LastName     string
	FirstName    string
	MiddleName   *string
	Email        string
	PasswordHash string
}

type UserUpdate struct {
	ID         string
	LastName   string
	FirstName  string
	MiddleName *string
}

type UserProfile struct {
	ID             string
	LastName       string
	FirstName      string
	MiddleName     *string
	Email          string
	EmailConfirmed bool
}
