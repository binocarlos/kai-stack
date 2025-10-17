package main

import (
	goapi "github.com/binocarlos/kai-stack/api/cmd/api"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	goapi.Execute()
}
