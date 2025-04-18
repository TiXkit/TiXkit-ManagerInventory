package models

import "time"

type User struct {
	ID        int
	FirstName string
	LastName  string
	IsAdmin   bool
	Email     string
	Hash      string
	UpdatedAt time.Time
	CreatedAt time.Time
}
