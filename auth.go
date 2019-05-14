package main

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	jose "gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"
)

// User es un usuario del REST API
type User struct {
	Name  string
	Pwd   string
	Roles map[string]bool // conjunto de roles
}

// userOnce es para iniciar el almacén de usuarios solo la primera vez.
var usersOnce sync.Once
// Almacén global de usuarios.
var users map[string]User

// initUsers crea usuarios de prueba
func initUsers() {
	users = map[string]User{
		"u0": {"u0", "u0", map[string]bool{"Admin": true}},
		"u1": {"u1", "u1", map[string]bool{"Edit": true}},
		"u2": {"u2", "u2", map[string]bool{"Add": true}},
	}
}

// private y public key para funciones crypto.
var privateKey *rsa.PrivateKey
var publicKey *rsa.PublicKey
// keysOnce es para iniciar los keys solo la primera vez.
var keysOnce sync.Once

// initKeys genera public/private keys
func initKeys() {
	var err error
	// En un ambiente en producción, estos keys se almacenarían con estricta seguridad
	// en el servidor, para poder cifrar y descifrar los tokens aun cuando el server
	// ha sido reiniciado.
	privateKey, err = rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatalf("Generando crypto keys: %v", err)
	}
	publicKey = &privateKey.PublicKey
}

// Errores relacionados a credenciales de usuarios.
var errUserNotFound = fmt.Errorf("no existe el usuario")
var errBadPassword = fmt.Errorf("contraseña incorrecta")

// getToken obtiene un JWE si los credenciales son correctos. El JWE expira en
// el periodo de tiempo especificado por exp.
func getToken(uname, pwd string, exp time.Duration) (string, error) {
	// Iniciamos usuarios y keys solo la primera vez.
	usersOnce.Do(initUsers)
	keysOnce.Do(initKeys)
	// Obtenemos usuario
	u, ok := users[uname]
	if !ok {
		// Usuario no encontrado
		return "", errUserNotFound
	}
	// Verificamos contraseña
	if u.Pwd != pwd {
		// Contraseña incorrecta
		return "", errBadPassword
	}
	// Creamos un  signer usando RSASSA-PSS (SHA512) con el private key. Esto es 
	// para firmar el token, no para cifrarlo.
	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.PS512, Key: privateKey}, nil)
	if err != nil {
		return "", err
	}
	// Creamos un encrypter para cifrar el token.
	encrypter, err := jose.NewEncrypter(jose.A128GCM,
		jose.Recipient{Algorithm: jose.RSA_OAEP, Key: publicKey},
		(&jose.EncrypterOptions{}).WithType("JWT").WithContentType("JWT"))
	if err != nil {
		return "", err
	}
	// Creamos los claims 
	cl := jwt.Claims{
		Subject: uname,
		Issuer:  "Dude the Builder",
		// Expiración de 1 minuto
		Expiry:  jwt.NewNumericDate(time.Now().Add(exp)),
	}
	// A pesar que el token estará cifrado, no enviaremos la contraseña del usuario.
	u.Pwd = ""
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
	// Obtenemos el JWE si los credenciales son válidos.
	token, err := getToken(r.FormValue("username"), r.FormValue("password"), 1 * time.Minute)
	if err != nil {
		log.Printf("tokenHandler: %v", err)
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	// Enviamos el JWE
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	io.WriteString(w, token)
}

// verifyToken valida un JWE y devuelve los roles del usuario si la validación es exitosa.
func verifyToken(jwe string) (map[string]bool, error) {
	keysOnce.Do(initKeys)
	// Leemos el JWE
	tok, err := jwt.ParseSignedAndEncrypted(jwe)
	if err != nil {
		return nil, err
	}
	// Desciframos el JWE, obteniendo el JWS anidado.
	jws, err := tok.Decrypt(privateKey)
	if err != nil {
		return nil, err
	}
	// Extraemos los claims y el usuario del JWS.
	cl := jwt.Claims{}
	u := User{}
	if err := jws.Claims(publicKey, &cl, &u); err != nil {
		return nil, err
	}
	// Validamos si el JWS no ha expirado.
	err = cl.Validate(jwt.Expected{
		Time: time.Now(),
	})
	if err != nil {
		return nil, err
	}
	// Devolvemos los roles del usuario para determinar niveles de autorización.
	return u.Roles, nil
}

// authWrapper envuelve un Handler en otro Handler que maneja autorización.
func authWrapper(h http.Handler) http.Handler {
	// Usamos un literal de función convertida a http.HandlerFunc
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
			http.Error(w, "Header Authorization inválido", http.StatusBadRequest)
			return
		}
		// Verificamos JWE y extraemos roles
		roles, err := verifyToken(parts[1])
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		// Admins pueden acceso total.
		if roles["Admin"] {
			h.ServeHTTP(w, r)
			return
		}
		// Autorización según el método y roles.
		switch r.Method {
		case "POST":
			if roles["Add"] {
				h.ServeHTTP(w, r)
				return
			}
		case "PUT", "DELETE":
			if roles["Edit"] {
				h.ServeHTTP(w, r)
				return
			}
		}
		// Combinación de roles y métodos inválida, acceso denegado.
		http.Error(w, "Acceso denegado", http.StatusUnauthorized)
	})
}
