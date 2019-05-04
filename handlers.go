package main

import (
	"net/http"

	"github.com/jecolon/post"
)

// postsHandler selecciona un HandlerFunc basado en el verbo HTTP.
func postsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		getPost(w, r)
	case "POST":
		newPost(w, r)
	case "PUT":
		putPost(w, r)
	case "DELETE":
		delPost(w, r)
	default:
		http.Error(w, "Método inválido", http.StatusMethodNotAllowed)
	}
}

// listPosts envía un listado de todos los posts.
func listPosts(w http.ResponseWriter, r *http.Request) {
	// Enviamos respuesta codificada como JSON
	sendJSON(w, post.List())
}

// getPost envía un post seún el ID del Request.
func getPost(w http.ResponseWriter, r *http.Request) {
	// Si no hay ID, enviamos listado.
	if r.URL.Path == "" {
		listPosts(w, r)
		return
	}
	// Obtenemos el post.
	var p post.Post
	if err := postFromRequest(w, r, &p); err != nil {
		return // Ya postFromRequest envío error al cliente.
	}
	// Enviamos respuesta codificada como JSON
	sendJSON(w, p)
}

// newPost crea un Post con los campos según el Request.Body en formato JSON.
func newPost(w http.ResponseWriter, r *http.Request) {
	// Recibimos el post como JSON y descodificamos
	var p post.Post
	if err := postFromJSON(w, r, &p); err != nil {
		return // Ya fromJSON evió error al cliente.
	}
	// Guardamos el post.
	p = post.New(p)[0]
	// Enviamos resultado codificado en JSON
	sendJSON(w, p)
}

// putPost actualiza un Post con los campos según el Request.Body en formato JSON.
func putPost(w http.ResponseWriter, r *http.Request) {
	// Verificamos que el post existe.
	var p post.Post
	if err := postFromRequest(w, r, &p); err != nil {
		return // Ya postFromRequest envío error al cliente.
	}
	// Guardamos id origianl
	id := p.Id
	// Descodificamos JSON para obtener campos actualizados.
	if err := postFromJSON(w, r, &p); err != nil {
		return // Ya fromJSON envío error al cliente.
	}
	// Asignamos ID original para evitar cambiar otro post.
	p.Id = id
	// Actualizamos el post.
	post.Put(p)
	// Enviamos respuesta codificada como JSON
	sendJSON(w, p)
}

// delPost borra un Post según el ID del Request.
func delPost(w http.ResponseWriter, r *http.Request) {
	// Verificamos que el post existe.
	var p post.Post
	if err := postFromRequest(w, r, &p); err != nil {
		return // Ya postFromRequest envío error al cliente.
	}
	// Borramos el post.
	post.Del(p.Id)
	// Enviamos objeto JSON vacío, señalando que se borró.
	sendJSON(w, struct{}{})
}
