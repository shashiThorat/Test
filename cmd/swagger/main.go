package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	// Run swag init command
	cmd := exec.Command("swag", "init", "-g", "pkg/api/server.go", "-o", "docs")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Println("Generating Swagger documentation...")
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Error generating Swagger documentation: %v\n", err)
		fmt.Println("Make sure swag is installed. Run: go install github.com/swaggo/swag/cmd/swag@latest")
		os.Exit(1)
	}

	fmt.Println("Swagger documentation generated successfully!")
	fmt.Println("Start the server and visit http://localhost:8080/swagger/index.html to view the API documentation.")
}