package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golangcollege/sessions"
)

type application struct {
	errorLog *log.Logger
	infoLog  *log.Logger
	session  *sessions.Session
}

func main() {
	port := flag.String("p", "3200", "HTTP network port")
	secret := flag.String("secret", "S4fcFbWc5caesR3d6ddSbGxvyzy31IIf", "App secret key")
	flag.Parse()

	infoLog := log.New(os.Stdout, "INFO:\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stdout, "ERROR:\t", log.Ldate|log.Ltime|log.Lshortfile)

	session := sessions.New([]byte(*secret))
	session.Lifetime = 48 * time.Hour

	app := &application{
		errorLog: errorLog,
		infoLog:  infoLog,
		session:  session,
	}

	srv := &http.Server{
		Addr:         "localhost:" + *port,
		ErrorLog:     errorLog,
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	infoLog.Printf("Starting server at http://localhost:%s\n", *port)
	err := srv.ListenAndServe()
	errorLog.Fatal(err)
}
