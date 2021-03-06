package main

import (
	"encoding/gob"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/charlesharries/podcast-stats/pkg/models"
	"github.com/charlesharries/podcast-stats/pkg/mysqlcache"
	"github.com/golangcollege/sessions"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/joho/godotenv"
)

type application struct {
	cache         *mysqlcache.Model
	errorLog      *log.Logger
	infoLog       *log.Logger
	episodes      *models.EpisodeModel
	listens       *models.ListenModel
	podcasts      *models.PodcastModel
	session       *sessions.Session
	subscriptions *models.SubscriptionModel
	templateCache map[string]*template.Template
	users         *models.UserModel
}

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil && os.Getenv("DB_NAME") == "" {
		log.Fatal("Couldn't load .env file.")
	}

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
	session := sessions.New([]byte(os.Getenv("APP_SECRET")))
	session.Lifetime = 48 * time.Hour
	gob.Register(TemplateUser{})

	// Assemble our application struct
	app := &application{
		cache:         &mysqlcache.Model{DB: db, Expiry: 24 * time.Hour},
		errorLog:      errorLog,
		infoLog:       infoLog,
		episodes:      &models.EpisodeModel{DB: db},
		listens:       &models.ListenModel{DB: db},
		podcasts:      &models.PodcastModel{DB: db},
		session:       session,
		subscriptions: &models.SubscriptionModel{DB: db},
		templateCache: templateCache,
		users:         &models.UserModel{DB: db},
	}

	// Create a custom server.
	srv := &http.Server{
		Addr:         os.Getenv("APP_HOST") + ":" + os.Getenv("PORT"),
		ErrorLog:     errorLog,
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Spin 'er up.
	infoLog.Printf("Starting server at http://localhost:%s\n", os.Getenv("PORT"))
	err = srv.ListenAndServe()
	errorLog.Fatal(err)
}

func openDB() (*gorm.DB, error) {

	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s)/%s?charset=utf8mb4,utf8&parseTime=true",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_NAME"),
	)
	db, err := gorm.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	// Test connection.
	if err := db.DB().Ping(); err != nil {
		return nil, err
	}

	db.AutoMigrate(
		&models.Episode{},
		&models.Listen{},
		&models.Podcast{},
		&models.Subscription{},
		&models.User{},
		&mysqlcache.CacheEntry{},
	)

	return db, nil
}
