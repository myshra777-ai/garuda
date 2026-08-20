package domain

import "time"

type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type Order struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Total     int64     `json:"total"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type UserRepository interface {
	GetUser(id string) (*User, error)
	CreateUser(u *User) error
}

type OrderRepository interface {
	SaveOrder(o *Order) error
	GetOrderByID(id string) (*Order, error)
}
