package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/ageapps/gambercoin/pkg/http_server"
	"github.com/rs/cors"
)

// our main function
func main() {

	var UIPort = flag.String("port", "8080", "Port for the UI client")
	flag.Parse()
	port, ok := os.LookupEnv("SERVER_PORT")
	if ok {
		UIPort = &port
	}
	router := http_server.NewRouter(false)
	log.Println("Listening on port: " + *UIPort)
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
	})

	handler := c.Handler(router)

	log.Fatal(http.ListenAndServe(":"+*UIPort, handler))
}
