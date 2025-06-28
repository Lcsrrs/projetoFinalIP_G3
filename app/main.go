package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"time" // Adicionado para lidar com datas

	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type dados_pessoais struct {
	Ddd_telefone      int
	Telefone          int
	Logradouro        string
	Complemento       string
	Uf                string
	Municipio         string
	Cep               string
	Numero_residencia string
	Bairro            string
	Codigo_municipio  string
	Ponto_referencia  string
}

type paciente struct {
	Id                      int
	Cartao_sus              string
	Nome_completo           string
	Nome_mae                string
	Apelido                 string
	Cpf                     string
	Nacionalidade           string
	Data_nascimento         string
	Idade                   int
	Raca_cor                string
	Raca_cor_outro          string
	Escolaridade            string
	Dados_pessoais_paciente dados_pessoais
}

type dados_anamnese struct {
	ID                      int
	PacienteID              int
	NomePaciente            string
	CPF                     string
	MotivoExame             string
	FezPreventivo           string
	DetalhesPreventivo      string
	UsaDiu                  string
	Gravidez                string
	UsaAnticoncepcional     string
	UsaHormonioMenopausa    string
	FezRadioterapia         string
	DataUltimaMenstruacao   sql.NullTime
	SangramentoPosRelacao   string
	SangramentoPosMenopausa string
	DataRegistro            time.Time
	NumeroProtocolo         string
}

var db = conexaoBanco()
var nome_usuario string
var tpl *template.Template
var resultado_busca []paciente
var cookie_sessao = sessions.NewCookieStore([]byte("super-secret"))

func main() {

	// Carregar o arquivo .env
	err := godotenv.Load("./app/.env")
	tpl, _ = template.ParseGlob("./static/*.html")

	fs := http.FileServer(http.Dir("./static"))

	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/", Autenticar(indexHandler))
	http.HandleFunc("/login", login)
	http.HandleFunc("/logout", logout)
	http.HandleFunc("/cadastro_usuario", cadastro_usuario)
	http.HandleFunc("/cadastro_pacientes", Autenticar(cadastrar_paciente))
	http.HandleFunc("/consultar_paciente", Autenticar(consultar_paciente))
	http.HandleFunc("/exame_clinico", Autenticar(registrar_exame_clinico))
	http.HandleFunc("/sucesso", Autenticar(paginaSucesso))
	http.HandleFunc("/anamnese", Autenticar(anamnese))
	http.HandleFunc("/consultar_atendimento_previo", Autenticar(consultarAnamnese))

	log.Println("Server rodando na porta 8080")

	err = http.ListenAndServe(":8080", context.ClearHandler(http.DefaultServeMux))
	if err != nil {
		panic(err)
	}

}

// Função Autenticar
// Esta função é um middleware que verifica se o usuário está autenticado antes de permitir o acesso.
func Autenticar(HandlerFunc http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := godotenv.Load("./app/.env")
		if err != nil {
			log.Fatalf("Erro ao carregar .env")
		}
		token_sessao := os.Getenv("NOME_SESSAO")
		sessao, _ := cookie_sessao.Get(r, token_sessao)
		_, ok := sessao.Values["userID"]
		if !ok {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		HandlerFunc.ServeHTTP(w, r)
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tpl.ExecuteTemplate(w, "index.html", nome_usuario)
}

func paginaSucesso(w http.ResponseWriter, r *http.Request) {
	tpl.ExecuteTemplate(w, "sucesso.html", nil)
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

		log.Fatalf("Erro ao conectar à database")
	}

	//Criando as tabelas no banco de dados
	// Cria a tabela usuarios_clinica

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
		fmt.Println(err)
		log.Fatalf("Erro ao criar tabela usuario_clinica")
	}

	// Cria a tabela pacientes

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

	//cria a tabela exame_clinico
	_, err = database.Query(`CREATE TABLE IF NOT EXISTS exame_clinico (
		id SERIAL PRIMARY KEY,
		inspecao_colo VARCHAR(100),
		sinais_dst VARCHAR(10),
		data_coleta DATE,
		responsavel VARCHAR(100)
	)`)
	if err != nil {
		log.Fatalf("Erro ao criar tabela exame_clinico")
	}
	//cria a tabela anamnese

	_, err = database.Query(`CREATE TABLE IF NOT EXISTS anamnese (
    id SERIAL PRIMARY KEY,
        paciente_id INTEGER REFERENCES pacientes(id),
        numero_protocolo VARCHAR(20) NOT NULL,
        motivo_exame VARCHAR(100) NOT NULL,
        fez_preventivo VARCHAR(20) NOT NULL,
        detalhes_preventivo TEXT,
        usa_diu VARCHAR(20) NOT NULL,
        gravidez VARCHAR(20) NOT NULL,
        usa_anticoncepcional VARCHAR(20) NOT NULL,
        usa_hormonio_menopausa VARCHAR(20) NOT NULL,
        fez_radioterapia VARCHAR(20) NOT NULL,
        data_ultima_menstruacao DATE,
        sangramento_pos_relacao VARCHAR(50) NOT NULL,
        sangramento_pos_menopausa VARCHAR(50) NOT NULL,
        data_registro TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)`)
	if err != nil {
		log.Fatalf("Erro ao criar tabela anamnese: %v", err)
	}

	return database

}

// Função para cadastrar um usuário
// Esta função é responsável por lidar com o cadastro de novos usuários na clínica
// Ela verifica se o método da requisição é GET ou POST, processa os dados do formulário
// e insere as informações no banco de dados, além de lidar com erros comuns como senhas não correspondentes
// e usuários já cadastrados.

func cadastro_usuario(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tpl.ExecuteTemplate(w, "cadastro_usuario.html", nil)
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

		// insere a senha hasheado no banco de dados
		senha_hasheada, _ := hashearSenha(senha)

		// Inserir os dados do usuário no banco de dados
		// A consulta SQL insere os dados do usuário na tabela usuarios_clinica

		_, err = db.Exec("INSERT INTO usuarios_clinica (CPF, email, nome_completo, cns, cnes, senha) VALUES ($1, $2, $3, $4, $5, $6)", CPF, email, nome_completo, CNS, CNES, senha_hasheada)
		if err != nil {
			http.Error(w, "Erro na inserção dos dados no banco de dados", http.StatusInternalServerError)
			fmt.Println("Erro: %v", err)
			return
		}

		http.Redirect(w, r, "/sucesso", http.StatusSeeOther)
	}

}

// Funções auxiliares para hashear e verificar senhas
// Essas funções utilizam o pacote bcrypt para gerar um hash seguro da senha do usuário
func hashearSenha(senha string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(senha), 10)
	return string(bytes), err
}

func checarSenhaHash(senha, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(senha))
	return err == nil
}

// Funções de login e logout
// Essas funções lidam com o processo de autenticação do usuário, verificando as credenciais.

func login(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tpl.ExecuteTemplate(w, "login.html", nil)
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

		var userID string

		linha := db.QueryRow("SELECT id, nome_completo FROM usuarios_clinica WHERE CPF = $1", cpf_usuario)
		err = linha.Scan(&userID, &nome_usuario)
		if err != nil {
			fmt.Println("Erro ao atribuir varíaveis")
		}
		if err == nil {
			sessao, _ := cookie_sessao.Get(r, "session")
			sessao.Values["userID"] = userID
			sessao.Save(r, w)
			http.Redirect(w, r, "/", http.StatusSeeOther)
		}
	}
}

func logout(w http.ResponseWriter, r *http.Request) {
	sessao, _ := cookie_sessao.Get(r, "session")
	delete(sessao.Values, "userID")
	sessao.Save(r, w)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// Funções para cadastro de pacientes e registro de exames clínicos
// Essas funções lidam com o cadastro de novos pacientes e o registro de exames clínicos.
func cadastrar_paciente(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tpl.ExecuteTemplate(w, "cadastro_pacientes.html", nil)
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
		//Converter idade para SMALLINT
		//o tipo SMALLINT em bancos de dados geralmente armazene números menores
		//garante que a idade seja convertida de forma segura, tratando tanto os casos em que a idade não é informada
		//e deve ser nula quanto os casos em que é informada, mas em um formato inválido.
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

		// Inserir dados na tabela pacientes
		// A consulta SQL insere os dados do paciente na tabela pacientes
		_, err := db.Exec(`INSERT INTO pacientes (
		cartao_sus, nome_completo_mulher, nome_completo_mae, apelido_mulher, cpf, 
		nacionalidade, data_nascimento, idade, raca_cor, raca_cor_outro, 
		ddd, telefone, escolaridade, logradouro, complemento, uf, 
		municipio, cep, numero_residencia, bairro, codigo_municipio, ponto_referencia) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22)`,
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

		fmt.Println("Paciente inserido com sucesso, redirecionando para /sucesso")
		http.Redirect(w, r, "/sucesso", http.StatusSeeOther) // Redirecionar após o sucesso (para a página de sucesso)
	}
}

// Função para registrar exame clínico
// Esta função é responsável por lidar com o registro de exames clínicos, recebendo os dados.

func registrar_exame_clinico(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tpl.ExecuteTemplate(w, "exame_clinico.html", nil)
		return
	} else if r.Method == http.MethodPost {
		inspecao_colo := r.FormValue("inspecao_colo")
		sinais_dst := r.FormValue("sinais_dst")
		data_coleta_str := r.FormValue("data_coleta")
		responsavel := r.FormValue("responsavel")

		//converter a data para o tipo de dado do SQL
		var data_coleta sql.NullTime
		if data_coleta_str != "" {
			parsedData, err := time.Parse("2006-01-02", data_coleta_str)
			if err != nil {
				http.Error(w, "Data inválida. Use o formato YYYY-MM-DD", http.StatusBadRequest)
				return
			}
			data_coleta = sql.NullTime{Time: parsedData, Valid: true}
		} else {
			data_coleta = sql.NullTime{Valid: false}
		}

		//Inserir dados na tabela (exame_clinico)
		_, err := db.Exec(`INSERT INTO exame_clinico
	 (inspecao_colo, sinais_dst, data_coleta, responsavel)
	  VALUES ($1, $2, $3, $4)`,
			inspecao_colo, sinais_dst, data_coleta, responsavel)

		if err != nil {
			log.Printf("Erro ao inserir dados exame clinico: %v", err)
			http.Error(w, "Erro ao registrar exame clínico. Verifique os dados e tente novamente", http.StatusInternalServerError)
			return
		}
		fmt.Println("Exame clínico registrado com sucesso!")
		http.Redirect(w, r, "/sucesso", http.StatusSeeOther)
	}
}

// Função para consultar pacientes
// Esta função lida com a consulta de pacientes, permitindo buscar por CPF, Cartão SUS ou nome completo.
// Ela exibe os resultados em uma página HTML, permitindo que o usuário veja os detalhes do paciente.
func consultar_paciente(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tpl.ExecuteTemplate(w, "consultar_paciente.html", resultado_busca)
		return
	} else if r.Method == http.MethodPost {
		resultado_busca = nil
		buscar := r.FormValue("campo_buscar")

		resultado, err := db.Query("SELECT id, cartao_sus, nome_completo_mulher, nome_completo_mae, apelido_mulher, cpf, nacionalidade, data_nascimento, idade, raca_cor, raca_cor_outro, escolaridade, ddd, telefone, logradouro, complemento, uf, municipio, cep, numero_residencia, bairro, codigo_municipio, ponto_referencia FROM pacientes WHERE cpf = $1 OR cartao_sus = $1 OR nome_completo_mulher ILIKE '%' || $1 || '%'", buscar)
		if err != nil {
			log.Printf("Erro na consulta SQL: %v", err)
			http.Error(w, "Erro ao buscar paciente", http.StatusInternalServerError)
			return
		}

		defer resultado.Close()

		for resultado.Next() {
			var busca_paciente paciente

			err := resultado.Scan(
				&busca_paciente.Id,
				&busca_paciente.Cartao_sus,
				&busca_paciente.Nome_completo,
				&busca_paciente.Nome_mae,
				&busca_paciente.Apelido,
				&busca_paciente.Cpf,
				&busca_paciente.Nacionalidade,
				&busca_paciente.Data_nascimento,
				&busca_paciente.Idade,
				&busca_paciente.Raca_cor,
				&busca_paciente.Raca_cor_outro,
				&busca_paciente.Escolaridade,
				&busca_paciente.Dados_pessoais_paciente.Ddd_telefone,
				&busca_paciente.Dados_pessoais_paciente.Telefone,
				&busca_paciente.Dados_pessoais_paciente.Logradouro,
				&busca_paciente.Dados_pessoais_paciente.Complemento,
				&busca_paciente.Dados_pessoais_paciente.Uf,
				&busca_paciente.Dados_pessoais_paciente.Municipio,
				&busca_paciente.Dados_pessoais_paciente.Cep,
				&busca_paciente.Dados_pessoais_paciente.Numero_residencia,
				&busca_paciente.Dados_pessoais_paciente.Bairro,
				&busca_paciente.Dados_pessoais_paciente.Codigo_municipio,
				&busca_paciente.Dados_pessoais_paciente.Ponto_referencia,
			)
			if err != nil {
				fmt.Println("Erro ao ler resultado, erro = ", err)
				return
			}

			fmt.Println("Busca executada com sucesso")
			resultado_busca = append(resultado_busca, busca_paciente)
		}

		http.Redirect(w, r, "/consultar_paciente", http.StatusSeeOther) // Redirecionar após o sucesso

	}
}

func consultar_atendimento_previo(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tpl.ExecuteTemplate(w, "consultar_atendimento_previo.html", nil)
		return
	}
}

// Função para registrar anamnese
// Esta função lida com o registro de anamneses, recebendo os dados do formulário
func anamnese(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tpl.ExecuteTemplate(w, "anamnese.html", nil)
		return
	} else if r.Method == http.MethodPost {

		numeroProtocolo := fmt.Sprintf("%d", time.Now().Unix())
		pacienteID := r.FormValue("paciente_id")
		motivoExame := r.FormValue("radioQ1")
		fezPreventivo := r.FormValue("radioQ2")
		detalhesPreventivo := r.FormValue("outro-input")
		usaDiu := r.FormValue("radioQ3")
		gravidez := r.FormValue("radioQ4")
		usaAnticoncepcional := r.FormValue("radioQ5")
		usaHormonioMenopausa := r.FormValue("radioQ6")
		fezRadioterapia := r.FormValue("radioQ7")
		dataUltimaMenstruacao := r.FormValue("date-input")
		sangramentoPosRelacao := r.FormValue("radioQ8")
		sangramentoPosMenopausa := r.FormValue("radioQ9")

		pacienteIDInt, err := strconv.Atoi(pacienteID)
		if err != nil {
			http.Error(w, "ID do paciente inválido", http.StatusBadRequest)
			return
		}

		var dataMenstruacao sql.NullTime
		if dataUltimaMenstruacao != "" {
			parsedDate, err := time.Parse("2006-01-02", dataUltimaMenstruacao)
			if err != nil {
				http.Error(w, "Formato de data inválido", http.StatusBadRequest)
				return
			}
			dataMenstruacao = sql.NullTime{Time: parsedDate, Valid: true}
		}

		_, err = db.Exec(`INSERT INTO anamnese (
            paciente_id, motivo_exame, fez_preventivo, detalhes_preventivo, 
            usa_diu, gravidez, usa_anticoncepcional, usa_hormonio_menopausa, 
            fez_radioterapia, data_ultima_menstruacao, 
            sangramento_pos_relacao, sangramento_pos_menopausa, numero_protocolo
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
			pacienteIDInt, motivoExame, fezPreventivo, detalhesPreventivo,
			usaDiu, gravidez, usaAnticoncepcional, usaHormonioMenopausa,
			fezRadioterapia, dataMenstruacao,
			sangramentoPosRelacao, sangramentoPosMenopausa, numeroProtocolo)

		if err != nil {
			log.Printf("Erro ao inserir anamnese: %v", err)
			http.Error(w, "Erro ao registrar anamnese", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/sucesso", http.StatusSeeOther)
	}
}

// Função para editar anamnese
// Esta função permite editar uma anamnese existente, recebendo os dados do formulário e atualizando o banco de dados.
func consultarAnamnese(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// pacienteID := r.URL.Query().Get("paciente_id")
		// if pacienteID == "" {
		// 	http.Error(w, "ID do paciente não fornecido", http.StatusBadRequest)
		// 	return
		// }

		// id, err := strconv.Atoi(pacienteID)
		// if err != nil {
		// 	http.Error(w, "ID do paciente inválido", http.StatusBadRequest)
		// 	return
		// }

		// rows, err := db.Query(`
		//     SELECT a.id, p.nome_completo_mulher, p.cpf,
		//            a.motivo_exame, a.data_registro, a.numero_protocolo
		//     FROM anamnese a
		//     JOIN pacientes p ON a.paciente_id = p.id
		//     WHERE a.paciente_id = $1
		//     ORDER BY a.data_registro DESC`, id)

		// if err != nil {
		// 	http.Error(w, "Erro ao consultar anamnese", http.StatusInternalServerError)
		// 	return
		// }
		// defer rows.Close()

		// var anamneses []dados_anamnese
		// for rows.Next() {
		// 	var a dados_anamnese
		// 	err := rows.Scan(
		// 		&a.ID, &a.NomePaciente, &a.CPF,
		// 		&a.MotivoExame, &a.DataRegistro, &a.NumeroProtocolo,
		// 	)
		// 	if err != nil {
		// 		http.Error(w, "Erro ao ler anamnese", http.StatusInternalServerError)
		// 		return
		// 	}
		// 	anamneses = append(anamneses, a)
		// }

		// data := struct {
		// 	NomePaciente string
		// 	CPF          string
		// 	Anamneses    []dados_anamnese
		// }{
		// 	NomePaciente: anamneses[0].NomePaciente,
		// 	CPF:          anamneses[0].CPF,
		// 	Anamneses:    anamneses,
		// }

		tpl.ExecuteTemplate(w, "consultar_atendimento_previo.html", nil)
	}
}
