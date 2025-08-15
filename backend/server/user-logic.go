package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"

	"github.com/golang-jwt/jwt/v5"
)

type Server struct{ DB *sql.DB }

type User struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"password,omitempty"`
	CreatedAt time.Time `json:"-"`
}

type userLogin struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type userRegister struct {
	Name     string `json:"name"     binding:"required,min=2"`
	Username string `json:"username" binding:"required,alphanum"`
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

func openDB() (*sql.DB, error) {
	_ = godotenv.Load(".env")
	dsn := "host=" + os.Getenv("DB_HOST") +
		" port=" + os.Getenv("DB_PORT") +
		" user=" + os.Getenv("DB_USER") +
		" password=" + os.Getenv("DB_PASS") +
		" dbname=" + os.Getenv("DB_NAME") +
		" sslmode=" + os.Getenv("SSL_MODE")
	db, err := sql.Open("postgres", dsn)
	log.Printf("%s", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

func (s *Server) registerHandler(c *gin.Context) {
	var in userRegister
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "hash failed"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	q := `
		INSERT INTO users (name, username, email, password_hash, created_at)
		VALUES ($1,$2,$3,$4, now())
		RETURNING id, created_at;
	`
	var id int64
	var created time.Time
	if err := s.DB.QueryRowContext(ctx, q, in.Name, in.Username, in.Email, string(hash)).Scan(&id, &created); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db insert failed"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"ok":       true,
		"userId":   id,
		"email":    in.Email,
		"loggedIn": false,
	})
}

func (s *Server) loginHandler(c *gin.Context) {
	var in userLogin
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	var id int64
	var name, username, email, passwordHash string
	q := `SELECT id, name, username, email, password_hash FROM users WHERE email=$1`
	err := s.DB.QueryRowContext(ctx, q, in.Email).Scan(&id, &name, &username, &email, &passwordHash)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(in.Password)) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	token, err := makeJWT(id, email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token error"})
		return
	}

	// אפשר לשמור ב-cookie HttpOnly/SameSite
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "auth",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		// Secure: true, // ב-HTTPS
		Expires: time.Now().Add(24 * time.Hour),
	})

	c.JSON(http.StatusOK, gin.H{
		"ok":       true,
		"userId":   id,
		"name":     name,
		"username": username,
		"email":    email,
	})
}

func makeJWT(userID int64, email string) (string, error) {
	secret := os.Getenv("JWT_SECRET")

	claims := jwt.MapClaims{
		"sub":   userID,
		"email": email,
		"exp":   time.Now().Add(24 * time.Hour).Unix(),
		"iat":   time.Now().Unix(),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(secret))
}
