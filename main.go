package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/jecolon/post"
)

func init() {
	// Creamos 10 posts para empezar.
	for i := 1; i <= 10; i++ {
		post.Add(post.Post{
			Id:     post.NewId(),
			UserId: 1,
			Title:  "sunt aut facere repellat provident occaecati excepturi optio reprehenderit",
			Body:   "quia et suscipit\nsuscipit recusandae consequuntur expedita et cum\nrepre",
		})
	}
}

func main() {
	// Definimos el puerto
	puerto := ":8443"

	// Creamos un Server con ajustes de seguridad.
	srv := &http.Server{
		Addr:           puerto,
		Handler:        nil, // DefaultServeMux
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MiB
	}
	// Aseguramos limpiar los recursos usados por el paquete post al hacer Shutdown.
	srv.RegisterOnShutdown(post.Shutdown)

	// Definimos rutas.
	webroot := "webroot"
	http.Handle("/", http.FileServer(http.Dir(webroot)))
	http.Handle("/api/v1/posts", http.StripPrefix("/api/v1/posts", http.HandlerFunc(postsHandler)))
	http.Handle("/api/v1/posts/", http.StripPrefix("/api/v1/posts/", http.HandlerFunc(postsHandler)))

	// Canal para señalar conexiones inactivas cerradas.
	conxCerradas := make(chan struct{})
	// Lanzamos goroutine para esperar señal y llamar Shutdown.
	go waitForShutdown(conxCerradas, srv)

	// Lanzamos el Server y estamos pendientes por si hay shut down.
	fmt.Printf("Servidor HTTPS en puerto %s listo. CTRL+C para detener.\n", puerto)
	// Archivos para certificado y key, generados por
	// /usr/local/go/src/crypto/tls/generate_cert.go --host localhost
	cert, key := "tls/cert.pem", "tls/key.pem"
	if err := srv.ListenAndServeTLS(cert, key); err != http.ErrServerClosed {
		// Error iniciando el Server. Posible conflicto de puerto, permisos, etc.
		log.Printf("Error durante ListenAndServeTLS: %v", err)
	}

	// Esperamos a que el shut down termine al cerrar todas las conexiones.
	<-conxCerradas
	fmt.Println("Shut down del servidor HTTPS completado exitosamente.")
}

// waitForShutdown para detectar señales de interrupción al proceso y hacer Shutdown.
func waitForShutdown(conxCerradas chan struct{}, srv *http.Server) {
	// Canal para recibir señal de interrupción.
	sigint := make(chan os.Signal, 1)
	// Escuchamos por una señal de interrupción del OS (SIGINT).
	signal.Notify(sigint, os.Interrupt)
	<-sigint

	// Si llegamos aquí, recibimos la señal, iniciamos shut down.
	// Noten se puede usar un Context para posible límite de tiempo.
	fmt.Println("\nShut down del servidor HTTPS iniciado...")
	// Límite de tiempo para el Shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		// Error aquí tiene que ser cerrando conexiones.
		log.Printf("Error durante Shutdown: %v", err)
	}

	// Cerramos el canal, señalando conexiones ya cerradas.
	close(conxCerradas)
}
