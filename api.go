package main

import (
	"context"
	"fmt"
	"net/http"
)

type ApiServer struct {
	addr  string
	store Storage
}

func NewApiServer(addr string, store Storage) *ApiServer {
	return &ApiServer{
		addr:  addr,
		store: store,
	}
}

/*
	func (s *ApiServer) Run() error {
		router := http.NewServeMux()
		router.HandleFunc("GET /users", s.GetAllUsersHandler)
		router.HandleFunc("GET /users/", s.GetUserByIdHandler)
		router.HandleFunc("POST /createUser", s.CreateUserHandler)

		server := http.Server{
			Addr:    s.addr,
			Handler: router,
		}
		fmt.Println("Server is running", server.Addr)

		return server.ListenAndServe()
	}
*/
func (s *ApiServer) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("authToken")
		if err != nil {
			http.Error(w, "Token is required", http.StatusUnauthorized)
			return
		}

		claims, err := ValidateToken(cookie.Value)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}
		userId, err := ExtractClaims(claims)
		if err != nil {
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "Id", userId)
		next(w, r.WithContext(ctx))

		//next(w, r)
	}
}

func (s *ApiServer) Run() error {

	router := http.NewServeMux()

	router.HandleFunc("/createItem", s.AuthMiddleware(
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPost {
				s.CreateItemHandler(w, r)
			} else {
				http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			}
		}))

	router.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			s.GetAllUsersHandler(w, r)
		} else {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})

	// Tek bir kullanıcıyı getirme (GET /users/{id})
	router.HandleFunc("/users/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			s.GetUserByIdHandler(w, r)
		} else {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})

	router.HandleFunc("/myItems", s.AuthMiddleware(
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet {
				s.GetItemsofUserHandler(w, r)
			} else {
				http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			}
		}))

	router.HandleFunc("/itemsByCreator", s.AuthMiddleware(
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet {
				s.GetItemsByCreatorHandler(w, r)
			} else {
				http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			}
		}))

	router.HandleFunc("/deleteItem/", s.AuthMiddleware(
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPatch {
				s.DeleteItemHandler(w, r)
			} else {
				http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			}
		}))
	router.HandleFunc("/changeItemStatus", s.AuthMiddleware(
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPut {
				s.ChangeItemStatusHandler(w, r)
			} else {
				http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			}
		}))

	// Yeni kullanıcı oluşturma (POST /register)
	router.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			s.CreateUserHandler(w, r)
		} else {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})

	// Yeni kullanıcı oluşturma (POST /login)
	router.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			s.LoginUserHandler(w, r)
		} else {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})
	router.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			s.LogoutUserHandler(w, r)
		} else {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})

	server := http.Server{
		Addr:    s.addr,
		Handler: router,
	}
	fmt.Println("Server is running", server.Addr)

	return server.ListenAndServe()
}
