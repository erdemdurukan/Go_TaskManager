package main

type User struct {
	Id       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}
type Item struct {
	Id          int    `json:"id"`
	CreatorId   int    `json:"creatorId"`
	ItemOwnerId int    `json:"itemOwnerId"`
	Status      string `json:"status"`
	Message     string `json:"message"`
	IsDeleted   bool   `json:"isDeleted"`
}
type CreateUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
type ItemStatusChangeRequest struct {
	Id     int    `json:"id"`
	Status string `json:"status"`
}
type CreateItemRequest struct {
	ItemOwnerId int    `json:"itemOwnerId"`
	Status      string `json:"status"`
	Message     string `json:"message"`
}

func NewUser(username, email, password string) *User {
	return &User{
		Username: username,
		Password: password,
		Email:    email,
	}
}
func NewItem(creatorId, itemOwnerId int,
	status string, message string) *Item {
	return &Item{
		CreatorId:   creatorId,
		ItemOwnerId: itemOwnerId,
		Status:      status,
		Message:     message,
	}
}
