package main

import (
	"database/sql"
	"flag"
	_ "github.com/go-sql-driver/mysql"
	"log/slog"
	"net/http"
	"os"
	"snippetbox/pkg/models/mysql"
)

type application struct {
	logger   *slog.Logger
	snippets *mysql.SnippetModel
}

func main() {

	addr := flag.String("addr", ":4000", "HTTP network address")
	dsn := flag.String("dsn", "root:admin@tcp(127.0.0.1:3306)/snippetbox?parseTime=true", "MySQL data source name)")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	db, err := openDB(*dsn)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	defer db.Close()

	app := &application{
		logger:   logger,
		snippets: &mysql.SnippetModel{DB: db},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", app.home)                        //Display the home page
	mux.HandleFunc("/snippet", app.showSnippet)          //Display a specific snippet
	mux.HandleFunc("/snippet/create", app.createSnippet) //Create a new snippet

	fileServer := http.FileServer(http.Dir("./ui/static/"))

	// "/static" prefix before the request reaches the file server.
	mux.Handle("/static/", http.StripPrefix("/static", fileServer)) //Serve a specific static file

	srv := &http.Server{
		Addr:    *addr,
		Handler: app.routes(),
	}

	logger.Info("Starting server on port", "address", *addr)
	err = srv.ListenAndServe()
	logger.Error(err.Error())
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)

	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}
