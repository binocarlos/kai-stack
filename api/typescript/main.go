package main

import (
	"flag"

	"github.com/binocarlos/kai-stack/api/pkg/types"
	"github.com/tkrajina/typescriptify-golang-structs/typescriptify"
)

func main() {
	filePath := flag.String("output", "../../frontend/src/types/gotypes.ts", "Output file path for TypeScript definitions")
	flag.Parse()

	converter := typescriptify.New().
		Add(types.Comic{}).
		AddEnum(types.AllComicTypes).
		Add(types.LoginRequest{}).
		Add(types.LoginResponse{}).
		Add(types.UserStatusResponse{}).
		Add(types.User{})
	converter.CreateInterface = true
	converter.BackupDir = ""
	err := converter.ConvertToFile(*filePath)
	if err != nil {
		panic(err.Error())
	}
}
