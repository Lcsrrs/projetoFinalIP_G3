package main

import (
	"crypto/rand"
	"encoding/base64"
	"log"
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

func gerarTokenDeSessao(tamanho int) string {
	bytes := make([]byte, tamanho)
	if _, err := rand.Read(bytes);err != nil {
		log.Fatalf("Erro ao gerar token de sessão: %v", err)
	}

	return base64.URLEncoding.EncodeToString(bytes)
}
