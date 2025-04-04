package main

import (
	"CueMind/internal/api"
	"net/http"
)

func main() {

	http.ListenAndServe(":8000", api.CreateEndpoints())
}
