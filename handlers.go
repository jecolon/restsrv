package main

import(
	"fmt"
	"net/http"

	"./post"
)

// postsHandler selecciona un HandlerFunc basado en el verbo HTTP.
func postsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		getPost(w, r)
	case "POST":
		addPost(w, r)
	case "PUT":
		setPost(w, r)
	case "DELETE":
		delPost(w, r)
	default:
		http.Error(w, "Método inválido", http.StatusMethodNotAllowed)
	}
}

// listPosts envía un listado de todos los posts.
func listPosts(w http.ResponseWriter, r *http.Request) {
	// Enviamos respuesta codificada como JSON
	envíaJSON(w, post.List())
}

// getPost envía un post seún el ID del Request.
func getPost(w http.ResponseWriter, r *http.Request) {
	// Si no hay ID, enviamos listado.
	if r.URL.Path == "" {
		listPosts(w, r)
		return
	}
	// Obtenemos ID.
	id, err := idFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Buscamos el post
	p, ok := post.Get(id)
	if !ok {
		// No encontramos ese ID
		http.Error(w, fmt.Sprintf("ID no encontrado: %d", id), http.StatusNotFound)
		return
	}
	// Enviamos respuesta codificada como JSON
	envíaJSON(w, p)
}

// addPost crea un Post con los campos según el Request.Body en formato JSON.
func addPost(w http.ResponseWriter, r *http.Request) {
	// Recibimos el post como JSON y descodificamos
	var p post.Post
	err := recibeJSON(r, &p)
	if err != nil {
		http.Error(w, "Error en formato JSON", http.StatusBadRequest)
		return
	}
	// Asignamos nuevo ID único.
	p.Id = post.NewId()
	// Guardamos el post.
	post.Add(p)
	// Enviamos resultado codificado en JSON
	envíaJSON(w, p)
}

// setPost actualiza un Post con los campos según el Request.Body en formato JSON.
func setPost(w http.ResponseWriter, r *http.Request) {
	// Obtenemos el ID
	id, err := idFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Verificamos que el post existe.
	if _, ok := post.Get(id); !ok {
		// No encontramos ese ID
		http.Error(w, fmt.Sprintf("ID no encontrado: %d", id), http.StatusNotFound)
		return
	}
	// Recibimos campos actualizados como JSON y descodificamos
	var p post.Post
	err = recibeJSON(r, &p)
	if err != nil {
		http.Error(w, "Error en formato JSON", http.StatusBadRequest)
		return
	}
	// Asignamos ID original para evitar cambiar otro post.
	p.Id = id
	// Actualizamod el post.
	post.Set(p)
	// Enviamos respuesta codificada como JSON
	envíaJSON(w, p)
}

// delPost borra un Post según el ID del Request.
func delPost(w http.ResponseWriter, r *http.Request) {
	// Obtenemos el ID
	id, err := idFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Verificamos que el post existe.
	if _, ok := post.Get(id); !ok {
		// No encontramos ese ID
		http.Error(w, fmt.Sprintf("ID no encontrado: %d", id), http.StatusNotFound)
		return
	}
	// Borramos el post.
	post.Del(id)
	// Enviamos objeto JSON vacío, señalando que se borró.
	envíaJSON(w, struct{}{})
}

