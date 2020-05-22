package main

import (
	"encoding/gob"
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/charlesharries/podcast-stats/pkg/models"
	"github.com/golangcollege/sessions"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

type application struct {
	errorLog      *log.Logger
	infoLog       *log.Logger
	session       *sessions.Session
	templateCache map[string]*template.Template
	users         *models.UserModel
}

func main() {
	port := flag.String("p", "3200", "HTTP network port")
	secret := flag.String("secret", "S4fcFbWc5caesR3d6ddSbGxvyzy31IIf", "App secret key")
	flag.Parse()

	// Create custom loggers.
	infoLog := log.New(os.Stdout, "INFO:\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stdout, "ERROR:\t", log.Ldate|log.Ltime|log.Lshortfile)

	// Open connection to the database.
	db, err := openDB()
	if err != nil {
		errorLog.Fatal(err)
	}
	defer db.Close()

	// Compile our templates.
	templateCache, err := newTemplateCache("./web/template")
	if err != nil {
		errorLog.Fatal(err)
	}

	// Create a new session.
	session := sessions.New([]byte(*secret))
	session.Lifetime = 48 * time.Hour
	gob.Register(TemplateUser{})

	// Assemble our application struct
	app := &application{
		errorLog:      errorLog,
		infoLog:       infoLog,
		session:       session,
		templateCache: templateCache,
		users:         &models.UserModel{DB: db},
	}

	// Create a custom server.
	srv := &http.Server{
		Addr:         "localhost:" + *port,
		ErrorLog:     errorLog,
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Spin 'er up.
	infoLog.Printf("Starting server at http://localhost:%s\n", *port)
	err = srv.ListenAndServe()
	errorLog.Fatal(err)
}

func openDB() (*gorm.DB, error) {
	db, err := gorm.Open("mysql", "root:@/podcast_stats?charset=utf8&parseTime=true")
	if err != nil {
		return nil, err
	}

	// Test connection.
	if err := db.DB().Ping(); err != nil {
		return nil, err
	}

	// Migrate schema.
	db.AutoMigrate(&models.User{})

	return db, nil
}
