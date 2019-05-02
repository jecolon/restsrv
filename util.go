package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/jecolon/post"
)

// postFromRequest obtiene el post solicitado en el http.Request. Si no existe,
// se envía error al cliente y devuelve un error, en cuyo caso p no se debe usar.
func postFromRequest(w http.ResponseWriter, r *http.Request, p *post.Post) error {
	// Convertimos el id a int.
	id64, err := strconv.ParseInt(r.URL.Path, 10, 64)
	id := int(id64) // strconv devuelve int64, queremos int
	if err != nil {
		err = fmt.Errorf("ID del post inválido: %s", r.URL.Path)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	// Buscamos el post.
	var ok bool
	if *p, ok = post.Get(id); !ok {
		// No encontramos ese ID
		err = fmt.Errorf("ID no encontrado: %d", id)
		http.Error(w, err.Error(), http.StatusNotFound)
		return err
	}

	return nil
}

// fromJSON descodifica JSON del http.Request.Body. Si ocurre un error,
// se envía error al cliente y se devuelve un error, en cuyo caso x no se debe usar.
func postFromJSON(w http.ResponseWriter, r *http.Request, p *post.Post) error {
	err := json.NewDecoder(r.Body).Decode(p)
	r.Body.Close()
	if err != nil {
		log.Printf("fromJSON de %v: %v", *p, err)
		http.Error(w, "Error en formato JSON", http.StatusBadRequest)
		return err
	}
	return nil
}

// sendJSON envía la respuesta codificada como JSON.
func sendJSON(w http.ResponseWriter, x interface{}) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "\t") // Opcional. Formato legible para humanos.
	err := enc.Encode(x)
	if err != nil {
		log.Printf("enviaJSON de %v: %v", x, err)
		http.Error(w, "Error en formato JSON", http.StatusInternalServerError)
	}
}
