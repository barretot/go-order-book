package models

type User struct {
	Name  string
	Email string
}

type CreateUserResult struct {
	UserId   string
	WalletId string
}
