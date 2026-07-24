package main

import (
	"github.com/joho/godotenv"
	"softixa-solutions.com/studentify/cmd"
)

func main() {
	godotenv.Load()
	cmd.Server()
}
