package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/peeeeeem/go-chat.git/app"
)

func main() {
	server := &http.Server{
		Handler: router(),
		Addr:    ":3000",
	}

	go func() {
		fmt.Println(server.ListenAndServe())
	}()

	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, syscall.SIGTERM, syscall.SIGINT)
	<-terminate

	shutdown(server)
}

func router() *mux.Router {
	router := mux.NewRouter()

	router.PathPrefix("/public/").
		Handler(http.StripPrefix("/public/", http.FileServer(http.Dir("./static"))))

	router.HandleFunc("/ws",
		app.WebSocketHandler()).Methods(http.MethodGet)

	return router
}

func shutdown(sv *http.Server) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sv.Shutdown(ctx); err != nil {
		fmt.Printf("shutting down server %s", err)
	}

	fmt.Printf("server gracefully stopped")
}
