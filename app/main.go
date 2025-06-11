package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {

	// Criando um servidor
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	log.Println("Server rodando na porta 8080")

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}

}

// Realizando a conexão com o banco de dados
func conexaoBanco() *sql.DB {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Erro ao carregar arquivo .env")
	}

	usuarioBD := os.Getenv("USUARIO")
	senhaUsuario := os.Getenv("SENHA")
	nomeBD := os.Getenv("NOME_BANCO_DE_DADOS")
	dadosParaConexao := "user=" + usuarioBD + " dbname=" + nomeBD + " password=" + senhaUsuario + " host=localhost port=5432 sslmode=disable"
	database, err := sql.Open("postgres", dadosParaConexao)
	if err != nil {
		panic(err)
	}

	_, err = database.Query("CREATE TABLE IF NOT EXISTS usuario_clinica (ID SERIAL PRIMARY KEY, primeiro_nome VARCHAR(32) NOT NULL, sobrenome VARCHAR(60) NOT NULL, cns VARCHAR(15) NOT NULL, cnes VARCHAR(15) NOT NULL, senha VARCHAR(20) NOT NULL)")
	if err != nil {
		panic(err)
	}

	_, err = database.Query("CREATE TABLE IF NOT EXISTS pacientes (ID SERIAL PRIMARY KEY, cartaoSUS VARCHAR(30) UNIQUE NOT NULL, nome_completo VARCHAR(255) NOT NULL, nome_mae VARCHAR(255) NOT NULL, apelido VARCHAR(100), CPF VARCHAR(15) UNIQUE NOT NULL, nacionalidade VARCHAR(50), data_nascimento VARCHAR(10), raca VARCHAR(10), ddd CHAR(2), telefone VARCHAR(10), escolaridade VARCHAR(20)")
	if err != nil {
		panic(err)
	}

	return database

}
