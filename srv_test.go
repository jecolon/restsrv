package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/jecolon/post"
)

func TestGetToken(t *testing.T) {
	tests := []struct {
		uname string
		pwd   string
		err   error
	}{
		{"uz", "uz", errUserNotFound},
		{"u0", "uz", errBadPassword},
		{"u0", "u0", nil},
	}

	for _, test := range tests {
		_, err := getToken(test.uname, test.pwd, 5*time.Second)
		if err != test.err {
			t.Errorf("getToken(%s, %s) returned error %v, expected %v", test.uname, test.pwd, err, test.err)
		}
	}
}

func TestVerifyToken(t *testing.T) {
	jwe, err := getToken("u0", "u0", 1*time.Second)
	if err != nil {
		t.Fatalf(`getToken("u0", "u0") returned error %v`, err)
	}

	roles, err := verifyToken(jwe)
	if err != nil {
		t.Fatalf("verigyToken(jwe) error: %v", err)
	}
	if !roles["Admin"] {
		t.Fatal(`verigyToken(jwe) roles["Admin"] is false`)
	}

	time.Sleep(62 * time.Second)
	_, err = verifyToken(jwe)
	if err == nil {
		t.Fatal("verifyToken(jwe) should have detected expired token!")
	}
}

func TestPostsHandler(t *testing.T) {
	if err := post.Init(); err != nil {
		t.Fatalf("post.Init() error: %v", err)
	}

	tests := []struct {
		uname  string
		pwd    string
		method string
		url    string
		body   io.Reader
		status int
	}{
		// Admin
		{"u0", "u0", "LIST", "/api/v1/posts/", nil, http.StatusOK},
		{"u0", "u0", "GET", "/api/v1/posts/", nil, http.StatusOK},
		{"u0", "u0", "POST", "/api/v1/posts/", nil, http.StatusOK},
		{"u0", "u0", "PUT", "/api/v1/posts/", nil, http.StatusOK},
		{"u0", "u0", "DELETE", "/api/v1/posts/", nil, http.StatusOK},
		// Editor
		{"u1", "u1", "LIST", "/api/v1/posts/", nil, http.StatusOK},
		{"u1", "u1", "GET", "/api/v1/posts/", nil, http.StatusOK},
		{"u1", "u1", "POST", "/api/v1/posts/", nil, http.StatusUnauthorized},
		{"u1", "u1", "PUT", "/api/v1/posts/", nil, http.StatusOK},
		{"u1", "u1", "DELETE", "/api/v1/posts/", nil, http.StatusOK},
		// Adder
		{"u2", "u2", "LIST", "/api/v1/posts/", nil, http.StatusOK},
		{"u2", "u2", "GET", "/api/v1/posts/", nil, http.StatusOK},
		{"u2", "u2", "POST", "/api/v1/posts/", nil, http.StatusOK},
		{"u2", "u2", "PUT", "/api/v1/posts/", nil, http.StatusUnauthorized},
		{"u2", "u2", "DELETE", "/api/v1/posts/", nil, http.StatusUnauthorized},
	}

	for _, test := range tests {
		p := post.New(post.Post{
			UserId: 1,
			Title:  "El título",
			Body:   "El contenido. " + test.url,
		})
		id := strconv.Itoa(p[0].Id)

		if test.method != "LIST" && test.method != "POST" {
			test.url += id
		}
		if test.method == "POST" || test.method == "PUT" {
			test.body = strings.NewReader(`{
				"UserId": 1,
				"Title": "El título",
				"Body": "El contenido."
			}`)
		}
		if test.method == "LIST" {
			test.method = "GET"
		}

		r := httptest.NewRequest(test.method, test.url, test.body)
		r.Header.Set("Content-Type", "application/json")
		tok, _ := getToken(test.uname, test.pwd, 3*time.Second)
		r.Header.Set("Authorization", "Bearer "+tok)
		w := httptest.NewRecorder()
		h := http.StripPrefix("/api/v1/posts/", authWrapper(http.HandlerFunc(postsHandler)))
		h.ServeHTTP(w, r)

		resp := w.Result()

		// Borramos post creado en esta prueba.
		if test.method == "POST" && resp.StatusCode == 200 {
			var rp post.Post
			err := json.NewDecoder(resp.Body).Decode(&rp)
			if err != nil {
				t.Fatalf("error decoding PUT json response: %v", err)
			}
			resp.Body.Close()
			post.Del(rp.Id)
		}

		if resp.StatusCode != test.status {
			t.Errorf("\npostsHandler user: %s method: %s\nurl: %s\ngot status: %d ; wanted %d",
				test.uname, test.method, test.url, resp.StatusCode, test.status)
		}

		// Remove test post.
		post.Del(p[0].Id)
	}
}
