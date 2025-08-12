package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	_ "github.com/lib/pq"
)

type User struct {
	Name      string    `json:"name"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"-"`
}

type UserLog struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Server struct{ DB *sql.DB }

func (srv *Server) loginFunc(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")

	if r.Method == http.MethodOptions { // preflight
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Printf("Incoming %s %s", r.Method, r.URL.Path)

	var in UserLog
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}
	if in.Email == "" || in.Password == "" {
		http.Error(w, `{"error":"missing fields"}`, http.StatusBadRequest)
		return
	}

	// TODO: encrypt password

	q := `
	SELECT id, name, email
    FROM users
    WHERE email = $1 AND password_hash = $2;
    `

	res, err := srv.DB.Exec(q, in.Email, in.Password)

	if err != nil {
		log.Println("insert error:", err)
		http.Error(w, `{"error":"db insert failed"}`, http.StatusInternalServerError)
		return
	} else {
		log.Printf("LOGGED IN")
	}
	n, _ := res.RowsAffected()
	log.Printf("Rows affected: %d", n)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{
		"ok":       true,
		"email":    in.Email,
		"loggedIn": true,
	})

}

func (srv *Server) registerFunc(w http.ResponseWriter, r *http.Request) {
	// CORS
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")

	if r.Method == http.MethodOptions { // preflight
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Printf("Incoming %s %s", r.Method, r.URL.Path)

	var in User
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}
	if in.Name == "" || in.Username == "" || in.Email == "" || in.Password == "" {
		http.Error(w, `{"error":"missing fields"}`, http.StatusBadRequest)
		return
	}

	// TODO: hash password properly (bcrypt). For now:
	hashPassword := in.Password

	q := `
		INSERT INTO users (name, username, email, password_hash, created_at)
		VALUES ($1, $2, $3, $4, now())
	`
	res, err := srv.DB.Exec(q, in.Name, in.Username, in.Email, hashPassword)
	if err != nil {
		log.Println("insert error:", err)
		http.Error(w, `{"error":"db insert failed"}`, http.StatusInternalServerError)
		return
	}
	n, _ := res.RowsAffected()
	log.Printf("Rows affected: %d", n)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{
		"ok":       true,
		"email":    in.Email,
		"loggedIn": false,
	})
}

func main() {
	log.Println("Starting Server...")

	db, err := openDB()
	if err != nil {
		log.Fatal("DB open:", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatal("DB ping:", err)
	}

	srv := &Server{DB: db}
	http.HandleFunc("/api/register", srv.registerFunc)
	http.HandleFunc("/api/login", srv.loginFunc)

	log.Println("Listening on :5423")
	if err := http.ListenAndServe(":5423", nil); err != nil {
		log.Fatal(err)
	}

}

func openDB() (*sql.DB, error) {
	dsn :=
		"host=localhost" +
			" port=5432" +
			" user=postgres" +
			" password=1234" +
			" dbname=video-chat-usersDB" +
			" sslmode=disable"
	log.Println(dsn)
	return sql.Open("postgres", dsn)
}
