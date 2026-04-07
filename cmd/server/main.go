package main

import (
	httptransport "breakfast-system/internal/transport/http"
	"breakfast-system/internal/transport/ws"
)

func main() {
	broker := ws.NewBroker()
	router := httptransport.NewRouter(broker)
	router.Run(":8080")
}
