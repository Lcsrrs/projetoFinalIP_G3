package main

import (
	"golang.org/x/crypto/bcrypt"
)

func hashearSenha(senha string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(senha), 10)
	return string(bytes), err
}

func checarSenhaHash(senha, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(senha))
	return err == nil
}
