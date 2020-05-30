package main

import (
	"encoding/gob"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/charlesharries/podcast-stats/pkg/cache"
	"github.com/charlesharries/podcast-stats/pkg/models"
	"github.com/golangcollege/sessions"
	"github.com/gomodule/redigo/redis"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/joho/godotenv"
)

type application struct {
	cache         *cache.Model
	errorLog      *log.Logger
	infoLog       *log.Logger
	podcasts      *models.PodcastModel
	session       *sessions.Session
	subscriptions *models.SubscriptionModel
	templateCache map[string]*template.Template
	users         *models.UserModel
}

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
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

	// Open the connection to our Redis cache.
	conn, err := openRedis()
	if err != nil {
		errorLog.Fatal(err)
	}
	defer conn.Close()

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
		cache:         &cache.Model{Conn: conn},
		errorLog:      errorLog,
		infoLog:       infoLog,
		podcasts:      &models.PodcastModel{DB: db},
		session:       session,
		subscriptions: &models.SubscriptionModel{DB: db},
		templateCache: templateCache,
		users:         &models.UserModel{DB: db},
	}

	// Create a custom server.
	srv := &http.Server{
		Addr:         "localhost:" + os.Getenv("PORT"),
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
		&models.User{},
		&models.Subscription{},
		&models.Podcast{},
	)

	return db, nil
}

func openRedis() (redis.Conn, error) {
	conn, err := redis.Dial("tcp", os.Getenv("REDIS_HOST"))
	if err != nil {
		return nil, err
	}

	_, err = conn.Do("AUTH", os.Getenv("REDIS_AUTH"))
	if err != nil {
		return nil, err
	}

	return conn, nil
}
