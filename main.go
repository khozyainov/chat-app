package main

import (
	"chat-app/chat"
	"chat-app/config"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

var configuration config.Configuration
var serverHostName string

func init() {
	configuration = config.LoadConfigAndSetUpLogging()
	serverHostName = fmt.Sprintf("%s:%s", configuration.Hostname, strconv.Itoa(configuration.Port))
	log.Println("The serverHost url", serverHostName)
}

func main() {

	// websocket server
	server := chat.NewServer()
	go server.Listen()
	http.HandleFunc("/", handleLoginPage)
	http.ListenAndServe(serverHostName, nil)
}

func handleLoginPage(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method == "GET" {
		http.ServeFile(responseWriter, request, "login.html")
	} else {
		http.Error(responseWriter, http.StatusText(405), 405)
		return
	}
}
