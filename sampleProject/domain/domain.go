package domain

type User struct {
	ID   string
	Name string
}

type UserRepository interface {
	Get(id string) *User
}
