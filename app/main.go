package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type Login struct {
	senhaHasheada string
	tokenSessao   string
	tokenCSRF     string
}

var db = conexaoBanco()
var usuarios = map[string]Login{}

func main() {

	// Criando um servidor
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	http.HandleFunc("/login", login)
	http.HandleFunc("/cadastro_usuario", cadastro_usuario)

	log.Println("Server rodando na porta 8080")

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}

}

// Realizando a conexão com o banco de dados
func conexaoBanco() *sql.DB {
	err := godotenv.Load("./app/.env")
	if err != nil {
		log.Fatalf("Erro ao carregar .env")
	}

	usuarioBD := os.Getenv("USUARIO")
	senhaUsuario := os.Getenv("SENHA")
	nomeBD := os.Getenv("NOME_BANCO_DE_DADOS")
	dadosParaConexao := "user=" + usuarioBD + " dbname=" + nomeBD + " password=" + senhaUsuario + " host=localhost port=5432 sslmode=disable"
	database, err := sql.Open("postgres", dadosParaConexao)
	if err != nil {
		fmt.Print(database)
		log.Fatalf("Erro ao conectar à database")
	}

	_, err = database.Query("CREATE TABLE IF NOT EXISTS usuarios_clinica (ID SERIAL PRIMARY KEY, CPF VARCHAR(15) UNIQUE NOT NULL, email VARCHAR(100) UNIQUE NOT NULL, nome_completo VARCHAR(64) NOT NULL, cns VARCHAR(15) NOT NULL, cnes VARCHAR(15) NOT NULL, senha VARCHAR(20) NOT NULL)")
	if err != nil {
		log.Fatalf("Erro ao criar tabela usuario_clinica")
	}

	_, err = database.Query("CREATE TABLE IF NOT EXISTS pacientes (ID SERIAL PRIMARY KEY, cartaoSUS VARCHAR(30) UNIQUE NOT NULL, nome_completo VARCHAR(255) NOT NULL, nome_mae VARCHAR(255) NOT NULL, apelido VARCHAR(100), CPF VARCHAR(15) UNIQUE NOT NULL, nacionalidade VARCHAR(50), data_nascimento VARCHAR(10), raca VARCHAR(10), ddd CHAR(2), telefone VARCHAR(10), escolaridade VARCHAR(20))")
	if err != nil {
		log.Fatalf("Erro ao criar tabela pacientes")
	}

	return database

}

func cadastro_usuario(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		http.ServeFile(w, r, "./static/cadastro_usuario.html")
		return
	}

	if r.Method == http.MethodPost {

		CPF := r.FormValue("CPF")
		email := r.FormValue("email")
		nome_completo := r.FormValue("nome_completo")
		senha := r.FormValue("senha")
		senha_confirmada := r.FormValue("confirmar_senha")
		CNS := r.FormValue("CNS")
		CNES := r.FormValue("CNES")

		fmt.Println(CPF, email, nome_completo, senha, senha_confirmada, CNS, CNES)

		if senha != senha_confirmada {
			err := http.StatusNotAcceptable
			http.Error(w, "Confirmação de senha não é válida", err)
			return
		}

		//verificar se já existe a conta no banco de dados
		var existe string
		err := db.QueryRow("SELECT CPF FROM usuarios_clinica WHERE CPF = $1", CPF).Scan(&existe)
		if err != sql.ErrNoRows {
			http.Error(w, "Usuário já cadastrado", http.StatusConflict)
			return
		}

		senha_hasheada, _ := hashearSenha(senha)
		usuarios[CPF] = Login{
			senhaHasheada: senha_hasheada,
		}

		dados, err := db.Prepare("INSERT INTO usuarios_clinica (CPF, email, nome_completo, cns, cnes, senha) VALUES ($1, $2, $3, $4, $5, $6)")
		if err != nil {
			log.Fatal("Erro na inserção dos dados no banco de dados")
		}
		dados.Exec(CPF, email, nome_completo, CNS, CNES, senha_hasheada)
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		err := http.StatusMethodNotAllowed
		http.Error(w, "Tipo de solicitação inválida", err)
		return
	}

	nome_usuario := r.FormValue("usuario")
	senha := r.FormValue("senha")

	usuario, ok := usuarios[nome_usuario]
	if !ok || !checarSenhaHash(senha, usuario.senhaHasheada) {
		err := http.StatusUnauthorized
		http.Error(w, "Usuário ou senha inválidos", err)
		return
	}

	tokenDeSessao := gerarTokenDeSessao(32)
	tokenCSRF := gerarTokenDeSessao(32)

	http.SetCookie(w, &http.Cookie{
		Name:     "tokenCSRF",
		Value:    tokenCSRF,
		Expires:  time.Now().Add(1 * time.Hour),
		HttpOnly: false,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "tokenDeSessao",
		Value:    tokenDeSessao,
		Expires:  time.Now().Add(1 * time.Hour),
		HttpOnly: true,
	})

	usuario.tokenCSRF = tokenCSRF
	usuario.tokenSessao = tokenDeSessao
	usuarios[nome_usuario] = usuario

	fmt.Fprintln(w, "Logado com sucesso!")

}

func area_protegida(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		err := http.StatusMethodNotAllowed
		http.Error(w, "Tipo de solicitação inválida", err)
		return
	}

	if err := autorizar(r); err != nil {
		er := http.StatusUnauthorized
		http.Error(w, "Não Autorizado", er)
		return
	}

	nome_usuario := r.FormValue("nome_usuario")
	fmt.Fprintf(w, "Validação CSRF feita com sucesso! Bem vindo, %s", nome_usuario)
}

func logout(w http.ResponseWriter, r *http.Request) {
	if err := autorizar(r); err != nil {
		er := http.StatusUnauthorized
		http.Error(w, "Não autorizado", er)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "tokenDeSessao",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HttpOnly: true,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "tokenCSRF",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HttpOnly: false,
	})

	nome_usuario := r.FormValue("nome_usuario")
	usuario, _ := usuarios[nome_usuario]
	usuario.tokenSessao = ""
	usuario.tokenCSRF = ""
	usuarios[nome_usuario] = usuario

	fmt.Fprintln(w, "Deslogado com sucesso!")
}
