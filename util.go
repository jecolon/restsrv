package main

import(
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

// Obtenemos el ID de un post del http.Request.
func idFromRequest(r *http.Request) (id int, err error) {
	// Convertimos el id de string a int64 y lo validamos
	id64, err := strconv.ParseInt(r.URL.Path, 10, 64)
	id = int(id64) // strconv devuelve int64, queremos int
	if err != nil {
		return -1, fmt.Errorf("ID del post inválido: %s", r.URL.Path)
	}

	return id, nil
}

// Recibimos JSON en el body del http.Request
func recibeJSON(r *http.Request, x interface{}) error {
	err := json.NewDecoder(r.Body).Decode(x)
	r.Body.Close()
	if err != nil {
		log.Printf("recibeJSON para %v: %v", x, err)
	}
	return nil
}

// Enviamos la respuesta codificada como JSON
func envíaJSON(w http.ResponseWriter, x interface{}) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "\t") // Opcional. Formato legible para humanos.
	err := enc.Encode(x)
	if err != nil {
		log.Printf("enviaJSON de %v: %v", x, err)
		http.Error(w, "Error en formato JSON", http.StatusInternalServerError)
	}
}
