package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time" // Adicionado para lidar com datas

	"golang.org/x/crypto/bcrypt"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var db = conexaoBanco()

func main() {

	// Criando um servidor
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	http.HandleFunc("/login", login)
	http.HandleFunc("/cadastro_usuario", cadastro_usuario)
	http.HandleFunc("/cadastro_pacientes", cadastrar_paciente)

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

	_, err = database.Query(`CREATE TABLE IF NOT EXISTS usuarios_clinica (
	ID SERIAL PRIMARY KEY,
	CPF VARCHAR(15) UNIQUE NOT NULL, 
	email VARCHAR(100) UNIQUE NOT NULL,
	nome_completo VARCHAR(64) NOT NULL, 
	cns VARCHAR(15) NOT NULL, 
	cnes VARCHAR(15) NOT NULL,
	senha VARCHAR(500) NOT NULL
	)`)
	if err != nil {
		log.Fatalf("Erro ao criar tabela usuario_clinica")
	}

	_, err = database.Query(`CREATE TABLE IF NOT EXISTS pacientes (
	ID SERIAL PRIMARY KEY,
	cartao_sus VARCHAR(30) NOT NULL,
	nome_completo_mulher VARCHAR(255) NOT NULL,
	nome_completo_mae VARCHAR(255) NOT NULL,
	apelido_mulher VARCHAR(100),
	cpf VARCHAR(15) UNIQUE NOT NULL,
	nacionalidade VARCHAR(50),
	data_nascimento DATE,
	idade SMALLINT,
	raca_cor VARCHAR(50),
	raca_cor_outro VARCHAR(100),
	ddd CHAR(2),
	telefone VARCHAR(15), 
	escolaridade VARCHAR(50),
	logradouro VARCHAR(255),
	complemento VARCHAR(255),
	uf CHAR(2),
	municipio VARCHAR(100),
	cep VARCHAR(10),
	numero_residencia VARCHAR(20), 
	bairro VARCHAR(100),
	codigo_municipio VARCHAR(10),
	ponto_referencia TEXT
	)`)

	if err != nil {
		log.Fatalf("Erro ao criar tabela pacientes")
	}

	return database

}

func cadastro_usuario(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		http.ServeFile(w, r, "./static/cadastro_usuario.html")
		return
	} else if r.Method == http.MethodPost {

		CPF := r.FormValue("CPF")
		email := r.FormValue("email")
		nome_completo := r.FormValue("nome_completo")
		senha := r.FormValue("senha")
		senha_confirmada := r.FormValue("confirmar_senha")
		CNS := r.FormValue("CNS")
		CNES := r.FormValue("CNES")

		if senha != senha_confirmada {
			err := http.StatusNotAcceptable
			http.Error(w, "Confirmação de senha não é válida", err)
			return
		}

		//verificar se já existe a conta no banco de dados
		var existe string
		err := db.QueryRow("SELECT CPF FROM usuarios_clinica WHERE CPF = $1", CPF).Scan(&existe)
		if err == nil {
			http.Error(w, "Usuário já cadastrado", http.StatusConflict)
			return
		}
		if err != sql.ErrNoRows {
			http.Error(w, "Erro ao verificar existência do usuário", http.StatusInternalServerError)
			return
		}

		senha_hasheada, _ := hashearSenha(senha)

		_, err = db.Exec("INSERT INTO usuarios_clinica (CPF, email, nome_completo, cns, cnes, senha) VALUES ($1, $2, $3, $4, $5, $6)", CPF, email, nome_completo, CNS, CNES, senha_hasheada)
		if err != nil {
			http.Error(w, "Erro na inserção dos dados no banco de dados", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

}

func hashearSenha(senha string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(senha), 10)
	return string(bytes), err
}

func checarSenhaHash(senha, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(senha))
	return err == nil
}

func login(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		http.ServeFile(w, r, "./static/login.html")
	} else if r.Method == http.MethodPost {

		cpf_usuario := r.FormValue("CPF")
		senha := r.FormValue("senha")

		var senhaHasheada string

		err := db.QueryRow("SELECT senha FROM usuarios_clinica WHERE CPF = $1", cpf_usuario).Scan(&senhaHasheada)

		if err == sql.ErrNoRows {
			http.Error(w, "Usuário não encontrado", http.StatusNotFound)
			return
		}

		if !checarSenhaHash(senha, senhaHasheada) {
			http.Error(w, "Usuário ou senha inválidos", http.StatusUnauthorized)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

}

func cadastrar_paciente(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		http.ServeFile(w, r, "./static/cadastro_pacientes.html")
		return
	} else if r.Method == http.MethodPost {
		// Extrair dados do formulário
		cartao_sus := r.FormValue("cartao_sus")
		nome_completo_mulher := r.FormValue("nome_completo_mulher")
		nome_completo_mae := r.FormValue("nome_completo_mae")
		apelido_mulher := r.FormValue("apelido_mulher")
		cpf := r.FormValue("cpf")
		nacionalidade := r.FormValue("nacionalidade")
		data_nascimento_str := r.FormValue("data_nascimento")
		idade_str := r.FormValue("idade")
		raca_cor := r.FormValue("raca_cor")
		raca_cor_outro := r.FormValue("raca_cor_outro")
		ddd := r.FormValue("ddd")
		telefone := r.FormValue("telefone")
		escolaridade := r.FormValue("escolaridade")
		logradouro := r.FormValue("logradouro")
		complemento := r.FormValue("complemento")
		uf := r.FormValue("uf")
		municipio := r.FormValue("municipio")
		cep := r.FormValue("cep")
		numero_residencia := r.FormValue("numero_residencia")
		bairro := r.FormValue("bairro")
		codigo_municipio := r.FormValue("codigo_municipio")
		ponto_referencia := r.FormValue("ponto_referencia")

		// Converter data de nascimento para tipo Date do SQL
		var dataNascimento sql.NullTime
		if data_nascimento_str != "" {
			parsedDate, err := time.Parse("2006-01-02", data_nascimento_str) // Formato YYYY-MM-DD
			if err != nil {
				http.Error(w, "Formato de data de nascimento inválido", http.StatusBadRequest)
				return
			}
			dataNascimento = sql.NullTime{Time: parsedDate, Valid: true}
		} else {
			dataNascimento = sql.NullTime{Valid: false}
		}
		// Converter idade para SMALLINT
		var idade sql.NullInt64
		if idade_str != "" {
			idadeVal, err := strconv.Atoi(idade_str)
			if err != nil {
				http.Error(w, "Formato de idade inválido", http.StatusBadRequest)
				return
			}
			idade = sql.NullInt64{Int64: int64(idadeVal), Valid: true}
		} else {
			idade = sql.NullInt64{Valid: false}
		}

		// Inserir dados no banco de dados
		_, err := db.Exec(`INSERT INTO pacientes (
		cartao_sus, nome_completo_mulher, nome_completo_mae, apelido_mulher, cpf, 
		nacionalidade, data_nascimento, idade, raca_cor, raca_cor_outro, 
		ddd, telefone, escolaridade, logradouro, complemento, uf, 
		municipio, cep, numero_residencia, bairro, codigo_municipio, ponto_referencia
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22)`,
			cartao_sus, nome_completo_mulher, nome_completo_mae, apelido_mulher, cpf,
			nacionalidade, dataNascimento, idade, raca_cor, raca_cor_outro,
			ddd, telefone, escolaridade, logradouro, complemento, uf,
			municipio, cep, numero_residencia, bairro, codigo_municipio, ponto_referencia,
		)

		if err != nil {
			log.Printf("Erro ao inserir paciente: %v", err) // Log para depuração
			http.Error(w, "Erro ao cadastrar paciente. Verifique os dados e tente novamente.", http.StatusInternalServerError)
			return
		}

		fmt.Println("Dados inseridos no banco com sucesso!")
		http.Redirect(w, r, "/", http.StatusSeeOther) // Redirecionar após o sucesso
	}

}
