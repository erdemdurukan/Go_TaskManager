package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func (s *ApiServer) GetAllUsersHandler(w http.ResponseWriter, r *http.Request) {
	users, err := s.store.getAllUser()

	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	WriteJSON(w, http.StatusOK, users)
}
func (s *ApiServer) GetItemsByCreatorHandler(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value("Id").(int)
	if !ok {
		http.Error(w, "Could not retrieve user ID", http.StatusInternalServerError)
		return
	}
	items, err := s.store.getItemsByCreator(userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	WriteJSON(w, http.StatusOK, items)
}
func (s *ApiServer) GetItemsofUserHandler(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value("Id").(int)
	if !ok {
		http.Error(w, "Could not retrieve user ID", http.StatusInternalServerError)
		return
	}
	items, err := s.store.getItemsofUser(userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	WriteJSON(w, http.StatusOK, items)
}

func (s *ApiServer) ChangeItemStatusHandler(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value("Id").(int)
	if !ok {
		http.Error(w, "Could not retrieve user ID", http.StatusInternalServerError)
		return
	}

	Req := ItemStatusChangeRequest{}
	if err := json.NewDecoder(r.Body).Decode(&Req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	if err := s.store.changeItemStatus(Req.Id, Req.Status, userId); err != nil {
		http.Error(w, "Could not update status of item", http.StatusInternalServerError)
		return
	}
	WriteJSON(w, http.StatusOK, "status changed")
}

func (s *ApiServer) GetUserByIdHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/users/")
	id, err := strconv.Atoi(path)

	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	user, err := s.store.getUserById(id)
	if err != nil {
		http.Error(w, "Could not find user", http.StatusBadRequest)
		return
	}
	WriteJSON(w, http.StatusOK, user)

}
func (s *ApiServer) DeleteItemHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/deleteItem/")
	id, err := strconv.Atoi(path)

	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	userId, ok := r.Context().Value("Id").(int)
	if !ok {
		http.Error(w, "Could not retrieve user ID", http.StatusInternalServerError)
		return
	}
	if err := s.store.deleteItem(id, userId); err != nil {
		http.Error(w, "Could not delete item", http.StatusInternalServerError)
		return

	}
	WriteJSON(w, http.StatusOK, "item deleted")

}

func (s *ApiServer) CreateItemHandler(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value("Id").(int)
	if !ok {
		http.Error(w, "Could not retrieve user ID", http.StatusInternalServerError)
		return
	}

	createItemReq := CreateItemRequest{}
	if err := json.NewDecoder(r.Body).Decode(&createItemReq); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	item1 := NewItem(userId, createItemReq.ItemOwnerId,
		createItemReq.Status, createItemReq.Message)

	//save item to the db
	if err := s.store.createItem(item1); err != nil {
		http.Error(w, "Could not create item", http.StatusInternalServerError)
		return

	}
	WriteJSON(w, http.StatusOK, item1)

}

func (s *ApiServer) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	//register

	createUserReq := CreateUserRequest{}
	if err := json.NewDecoder(r.Body).Decode(&createUserReq); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(createUserReq.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Could not hash password", http.StatusInternalServerError)
		return
	}

	//create user with constructor
	user1 := NewUser(createUserReq.Username, createUserReq.Email, string(hashedPassword))

	//save user to the db
	if err := s.store.createUser(user1); err != nil {
		http.Error(w, "Could not create user", http.StatusInternalServerError)
		return
	}

	WriteJSON(w, http.StatusOK, user1)
}

// exit authToken cookie'sini sil
func (s *ApiServer) LogoutUserHandler(w http.ResponseWriter, r *http.Request) {
	// Cookie'den authToken'ı al
	cookie, err := r.Cookie("authToken")

	if err != nil {
		http.Error(w, "No token found", http.StatusUnauthorized)
		return
	}

	// Token'ı doğrula
	claims, err := ValidateToken(cookie.Value)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	userId, ok := (*claims)["Id"].(float64) // JSON'dan gelen int, float64 olarak döner
	if !ok {
		http.Error(w, "Invalid token claims", http.StatusUnauthorized)
		return
	}

	// Token geçerli ise, çıkış işlemi yapılabilir
	// Cookie'den authToken'ı sil
	http.SetCookie(w, &http.Cookie{
		Name:     "authToken",
		Value:    "",
		Path:     "/",
		Expires:  time.Now().Add(-1 * time.Hour), // Token'ı geçersiz kıl
		HttpOnly: true,
		Secure:   false, // HTTPS üzerinde çalışırken true yapın
	})

	// Kullanıcı bilgisi ile birlikte çıkış mesajı döndür
	WriteJSON(w, http.StatusOK, map[string]string{
		"message": fmt.Sprintf("User %v logged out successfully", int(userId)),
	})
}

func (s *ApiServer) LoginUserHandler(w http.ResponseWriter, r *http.Request) {
	// Gelen JSON'u ayrıştır
	loginUserReq := LoginRequest{}
	if err := json.NewDecoder(r.Body).Decode(&loginUserReq); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	user, err := s.store.FindByEmail(loginUserReq.Email) // Veritabanı işlemi
	if err != nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginUserReq.Password)); err != nil {
		http.Error(w, "Invalid password", http.StatusUnauthorized)
		return
	}

	// Token'ı oluştur
	token, err := GenerateToken(user.Id)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Token'ı HTTP-only cookie olarak set et
	http.SetCookie(w, &http.Cookie{
		Name:     "authToken",
		Value:    token,
		Path:     "/",
		HttpOnly: true,  // JavaScript erişimini engeller
		Secure:   false, // HTTPS üzerinde çalışırken true yapın
		MaxAge:   3600,  // 1 saat geçerli
	})

	// Başarı mesajı
	WriteJSON(w, http.StatusOK, map[string]string{"message": "Logged in successfully"})
}
