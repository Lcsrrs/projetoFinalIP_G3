package main

import (
	"errors"
	"net/http"
)

var erroAutorizacao = errors.New("Não autorizado")

func autorizar(r *http.Request) error {
	nome_usuario := r.FormValue("nome_usuario")
	usuario, ok := usuarios[nome_usuario]

	if !ok {
		return erroAutorizacao
	}

	st, err := r.Cookie("tokenDeSesao")
	if err != nil || st.Value == "" || st.Value != usuario.tokenSessao {
		return erroAutorizacao
	}

	csrf := r.Header.Get("X-CSRF-Token")
	if csrf != usuario.tokenCSRF || csrf == "" {
		return erroAutorizacao
	}

	return nil
}
