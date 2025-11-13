package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

type Curso struct {
	ID            int     `json:"id"`
	Curso         string  `json:"curso"`
	NotaMinimaAC  float64 `json:"notaMinimaAC"`
	NotaMinimaEP  float64 `json:"notaMinimaEP"`
	NotaMinimaPPI float64 `json:"notaMinimaPPI"`
}

var db *sql.DB

func main() {
	var err error

	connStr := fmt.Sprintf(
    "host=%s port=%s user=%s password=%s dbname=%s sslmode=require",
    os.Getenv("DB_HOST"),
    os.Getenv("DB_PORT"),
    os.Getenv("DB_USER"),
    os.Getenv("DB_PASSWORD"),
    os.Getenv("DB_NAME"),
)
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r := gin.Default()

	// Middleware simples de CORS
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Rotas
	r.GET("/cursos", getCursos)
	r.GET("/curso/:id", getCursoByID)

	r.Run(":" + os.Getenv("PORT"))
}

func getCursos(c *gin.Context) {
	rows, err := db.Query("SELECT id, curso FROM cursos ORDER BY curso")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var cursos []Curso
	for rows.Next() {
		var curso Curso
		if err := rows.Scan(&curso.ID, &curso.Curso); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		cursos = append(cursos, curso)
	}

	c.JSON(http.StatusOK, cursos)
}

func getCursoByID(c *gin.Context) {
	id := c.Param("id")

	var curso Curso
	err := db.QueryRow(`
		SELECT id, curso, "notaminimaac", "notaminimaep", "notaminimappi"
		FROM cursos WHERE id = $1
	`, id).Scan(&curso.ID, &curso.Curso, &curso.NotaMinimaAC, &curso.NotaMinimaEP, &curso.NotaMinimaPPI)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Curso não encontrado"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, curso)
}
