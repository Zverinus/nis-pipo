package transport

import (
	"encoding/json"
	"net/http"
	"nis-pipo/internal/user"
)

type UserHandler struct {
	service *user.Service
}

func NewUserHandler (service *user.Service) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler)InitHandling() {
	http.Handle("/api/users/create", h.Create()) // POST: username,email,password,first_name,last_name
	http.Handle("/api/users/get",    h.Get())    // GET:   id 
	http.Handle("/api/users/list",   h.List())   // GET:   
	
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	})
}

func enableCors(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func (h *UserHandler) Create() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		enableCors(w)
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed); return
		}

		var req struct {
			Username  string `json:"username"`
			Email     string `json:"email"`
			Password  string `json:"password"`
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
		}
		ct := r.Header.Get("Content-Type")
		if ct != "" && ct[:16] == "application/json" {
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "bad json", http.StatusBadRequest); 
				return
			}
		} else {
			_ = r.ParseForm()
			req.Username  = r.FormValue("username")
			req.Email     = r.FormValue("email")
			req.Password  = r.FormValue("password")
			req.FirstName = r.FormValue("first_name")
			req.LastName  = r.FormValue("last_name")
		}

		u, err := h.service.Create(r.Context(), req.Username, req.Email, req.Password, req.FirstName, req.LastName)
		if err != nil { 
			http.Error(w, "cannot create", http.StatusBadRequest); 
			return 
		}

		writeJSON(w, http.StatusCreated, map[string]any{
			"id": u.ID, "username": u.Username, "email": u.Email,
			"first_name": u.FirstName, "last_name": u.LastName,
			"created_at": u.CreatedAt, "updated_at": u.UpdatedAt,
		})
	})
}

func (h *UserHandler) Get() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		enableCors(w)
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed); return
		}
		id := r.URL.Query().Get("id")
		if id == "" { 
			http.Error(w, "id required", http.StatusUnprocessableEntity); 
			return 
	}

		u, err := h.service.GetByID(r.Context(), id)
		if err != nil { 
			http.Error(w, "not found", http.StatusNotFound); 
			return 
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"id": u.ID, "username": u.Username, "email": u.Email,
			"first_name": u.FirstName, "last_name": u.LastName,
			"created_at": u.CreatedAt, "updated_at": u.UpdatedAt,
		})
	})
}


func (h *UserHandler) List() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		enableCors(w)
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed); return
		}

		us, err := h.service.List(r.Context(), 50, 0)
		if err != nil { 
			http.Error(w, "error", http.StatusInternalServerError);
			return 
		}
		writeJSON(w, http.StatusOK, us)
	})
}