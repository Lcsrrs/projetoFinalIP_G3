
# Sistema de Cadastro e Acompanhamento de Exame Citopatológico

Este projeto foi desenvolvido como trabalho final da disciplina de **Introdução à Programação**, do curso de **Engenharia de Software**, em parceria com a **Faculdade de Farmácia**.

O objetivo do sistema é **digitalizar a ficha de exame citopatológico**, utilizada no processo de cadastro, anamnese e exame clínico de mulheres que estão no processo de **rastreamento, diagnóstico e acompanhamento do câncer de colo de útero**.

---

## ⚙️ Tecnologias Utilizadas

- **Linguagem:** Go (Golang)
- **Banco de Dados:** PostgreSQL
- **Gerenciamento de Sessões:** Gorilla Sessions
- **Manipulação de Templates HTML:** nativo do Go (`html/template`)
- **Gerenciamento de variáveis de ambiente:** [godotenv](https://github.com/joho/godotenv)

---

## 📑 Funcionalidades

- Cadastro de usuários da clínica (com autenticação e senha criptografada).
- Login e logout de usuários.
- Cadastro de pacientes.
- Consulta de pacientes por CPF, Cartão SUS ou nome.
- Registro de exame clínico.
- Registro de anamnese.
- Consulta de atendimento prévio.

---

## 🚀 Como Executar

### ✅ Pré-requisitos:

- Ter o [Go](https://golang.org/doc/install) instalado (versão 1.20 ou superior recomendada).
- Ter o [PostgreSQL](https://www.postgresql.org/) instalado e rodando localmente.
- Criar um banco de dados PostgreSQL para o projeto.

### 🔧 Variáveis de Ambiente

Crie um arquivo `.env` dentro da pasta `/app` com o seguinte conteúdo (exemplo):

```
USUARIO=seu_usuario_postgres
SENHA=sua_senha_postgres
NOME_BANCO_DE_DADOS=nome_do_banco
NOME_SESSAO=session
```

### 📦 Instalação de dependências

Dentro da raiz do projeto, execute:

```bash
go mod tidy
```

### ▶️ Rodando o projeto

Para iniciar o servidor, execute:

```bash
go run ./app/main.go
```

O servidor estará disponível em:

```
http://localhost:8080
```

---

Ao acessar o server local, acessar manualmente a rota:

```
http://localhost:8080/cadastro_usuario
```

e realizar o cadastro de um usuário para ser utilizado na página de login

---

## 🗄️ Estrutura do Projeto

```
/app
  ├── main.go          -> Arquivo principal (back-end)
  ├── .env             -> Variáveis de ambiente
/static                -> Páginas HTML (front-end) e arquivos estáticos
  ├── style            
    ├── #.css          -> Arquivos CSS (estilização)
  ├── #.html           -> Arquivos HTML (estruturação)
```

---

## 🏗️ Banco de Dados

Ao rodar o projeto pela primeira vez, ele cria automaticamente as tabelas necessárias, caso não existam:

- `usuarios_clinica` → Cadastro e autenticação de usuários.
- `pacientes` → Dados das pacientes.
- `exame_clinico` → Registro dos exames clínicos.
- `consultas` → Registro das consultas a serem feitas
- `anamnese` → Registro das anamneses realizadas

---

## 🚩 Observações

- Este projeto é acadêmico e não deve ser utilizado diretamente em ambiente de produção sem as devidas validações de segurança, autenticação robusta e tratamento de dados sensíveis.

---

## 🤝 Colaboração

Projeto desenvolvido pelos alunos do curso de **Engenharia de Software**, em conjunto com a **Faculdade de Farmácia**, como parte do projeto de digitalização de processos do rastreamento do câncer de colo de útero.

---
