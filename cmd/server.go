package cmd

import (
	"net/http"

	"softixa-solutions.com/studentify/internal/config"
	"softixa-solutions.com/studentify/internal/route"
)

func Server() {
	config.Connect()
	router := route.Router()
	http.ListenAndServe(":8080", router)
}
