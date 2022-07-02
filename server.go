package main

import (
	"flag"
	"log"
	"path/filepath"
	"net/http"
	"os"
	"os/signal"
	"time"
	"context"
	"14joined.me/cs/handlers"
	"14joined.me/cs/middleware"
)

var (
	addr = flag.String("address", "127.0.0.1:9000", "listen address")
	files = flag.String("public", "./public", "client files")
)

func main() {
	flag.Parse()

	err := run(*addr, *files)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Server gracefully shutdown")
}

func run(addr, files string) error {
	mux := http.NewServeMux()
	mux.Handle(
		"/static/",
		http.StripPrefix("/static/",
			middleware.RestrictPrefix(
				".", http.FileServer(http.Dir(files)),
			),
		),
	)

	mux.Handle(
		"/",
		handlers.Methods{
			http.MethodGet: http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					w.Write([]byte("hello, world!"))
				},
			),
		},
	)

	mux.Handle(
		"/login",
		handlers.Methods{
			http.MethodGet: http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					http.ServeFile(w, r, filepath.Join(files, "login.html"))
				},
			),
		},
	)

	mux.Handle(
		"/request-access",
		handlers.Methods{
			http.MethodPost: http.HandlerFunc(
				handlers.RequestAccess(),
			),
		},
	)

	srv := &http.Server{
		Addr: addr,
		Handler: mux,
		IdleTimeout: time.Minute,
		ReadHeaderTimeout: 30 * time.Second,
	}

	done := make(chan struct{})
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)

		for {
			if <-c == os.Interrupt {
				if err := srv.Shutdown(context.Background()); err != nil {
					log.Printf("Shutdown: %v", err)
				}
				close(done)
				return
			}
		}
	}()

	log.Printf("Serving files in %q over %s\n", files, srv.Addr)

	var err error
	log.Println("ListenAndServer...")
	err = srv.ListenAndServe()

	if err == http.ErrServerClosed {
		err = nil 
	}

	<-done

	return err 
}
