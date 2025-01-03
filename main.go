package main

import (
	"log"
	"net/http"
	"qerplunk/garin-chat/envconfig"
	"qerplunk/garin-chat/middleware"
	"qerplunk/garin-chat/ws_server"
)

func main() {
	if envConfig := envconfig.InitEnvConfig(); !envConfig {
		return
	}

	middlewareStack := middleware.CreateStack(
		middleware.OriginCheck(),
	)

	http.HandleFunc("/", middlewareStack(wsserver.HandleWebSocket))

	port := envconfig.EnvConfig.Port

	log.Printf("WebSocket server running on ws://localhost:%s/\n", port)
	if serveErr := http.ListenAndServe(":"+port, nil); serveErr != nil {
		log.Println("Error starting server:", serveErr)
	}
}
