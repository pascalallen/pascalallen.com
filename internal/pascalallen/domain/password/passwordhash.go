package password

import (
	"golang.org/x/crypto/bcrypt"
	"log"
)

type PasswordHash string

func Create(password string) PasswordHash {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}

	return PasswordHash(passwordHash)
}

func (p *PasswordHash) Compare(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(*p), []byte(password))
	if err != nil {
		return false
	}

	return true
}
