package main

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	jose "gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"
)

type User struct {
	Name  string
	Pwd   string
	Roles map[string]bool // conjunto de roles
}

// Almacén global de usuarios. Solo para leer, seguro para acceso por múltiples goroutines.
var users map[string]User
// private y public key para funciones crypto.
var privateKey *rsa.PrivateKey
var publicKey *rsa.PublicKey


func init() {
	// Generamos public/private keys
	var err error
	privateKey, err = rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatalf("Generando crypto keys: %v", err)
	}
	publicKey = &privateKey.PublicKey

	// Creamos usuarios de prueba
	users = map[string]User{
		"u0": {"u0", "u0", map[string]bool{"Admin": true}},
		"u1": {"u1", "u1", map[string]bool{"Edit": true}},
		"u2": {"u2", "u2", map[string]bool{"Add": true}},
	}
}

// getToken obtiene un JWE si los credenciales son correctos.
func getToken(uname, pwd string) (string, error) {
	// Obtenemos usuario
	u, ok := users[uname]
	if !ok {
		return "", fmt.Errorf("no existe el usuario: %s", uname)
	}
	// Verificamos contraseña
	if u.Pwd != pwd {
		return "", fmt.Errorf("contraseña incorrecta")
	}
	// Creamos un  signer usando RSASSA-PSS (SHA512) con el private key.
	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.PS512, Key: privateKey}, nil)
	if err != nil {
		return "", err
	}
	// Creamos un encrypter
	encrypter, err := jose.NewEncrypter(jose.A128GCM,
		jose.Recipient{Algorithm: jose.RSA_OAEP, Key: publicKey},
		(&jose.EncrypterOptions{}).WithType("JWT").WithContentType("JWT"))
	if err != nil {
		return "", err
	}
	// Creamos los claims con expiración de 1 minuto
	cl := jwt.Claims{
		Subject: "uname",
		Issuer:  "https://josecolon.dev",
		Expiry:  jwt.NewNumericDate(time.Now().Add(60 * time.Second)),
	}
	// Creamos el JWE incluyendo claims y el usuario
	jwe, err := jwt.SignedAndEncrypted(signer, encrypter).Claims(cl).Claims(u).CompactSerialize()
	if err != nil {
		return "", err
	}

	// Devolvemos resultado
	return jwe, nil
}

// tokenHandler produce un JWE si los credenciales son correctos.
func tokenHandler(w http.ResponseWriter, r *http.Request) {
	token, err := getToken(r.FormValue("username"), r.FormValue("password"))
	if err != nil {
		log.Printf("tokenHandler: %v", err)
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	io.WriteString(w, token)
}

// verifyToken valida un JWE y devuelve los roles del usuario si la validación es exitosa.
func verifyToken(jwe string) (map[string]bool, error) {
	tok, err := jwt.ParseSignedAndEncrypted(jwe)
	if err != nil {
		return nil, err
	}

	nested, err := tok.Decrypt(privateKey)
	if err != nil {
		return nil, err
	}

	cl := jwt.Claims{}
	u := User{}
	if err := nested.Claims(publicKey, &cl, &u); err != nil {
		return nil, err
	}

	err = cl.Validate(jwt.Expected{
		Time: time.Now(),
	})
	if err != nil {
		return nil, err
	}

	return u.Roles, nil
}

// authWrapper envuelve un Handler en otro Handler que maneja autorización.
func authWrapper(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Libre acceso a GET y LIST
		if r.Method == "GET" {
			h.ServeHTTP(w, r)
			return
		}
		// Extraemos JWE del header
		authHeader := r.Header.Get("Authorization")
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 {
			http.Error(w, "Invalid Authorization header", http.StatusUnauthorized)
			return
		}
		// Verificamos JWE y extraemos roles
		roles, err := verifyToken(parts[1])
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		// Admins pueden hacer de todo
		if _, ok := roles["Admin"]; ok {
			h.ServeHTTP(w, r)
			return
		}
		// Actuamos según el Method y el rol.
		switch r.Method {
		case "POST":
			if _, ok := roles["Add"]; !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
		case "PUT", "DELETE":
			if _, ok := roles["Edit"]; !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
		}
		// Rol adecuado existe, procedemos
		h.ServeHTTP(w, r)
	})
}
