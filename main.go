package main

import (
	"fmt"
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
		middleware.JWTCheck(),
		middleware.OriginCheck(),
	)

	http.HandleFunc("/", middlewareStack(wsserver.HandleWebSocket))

	port := envconfig.EnvConfig.Port

	fmt.Printf("WebSocket server running on ws://localhost:%s/\n", port)
	if serveErr := http.ListenAndServe(":"+port, nil); serveErr != nil {
		fmt.Println("Error starting server:", serveErr)
	}
}
