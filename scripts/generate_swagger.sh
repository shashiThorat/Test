#!/bin/bash

# Generate Swagger documentation
swag init -g pkg/api/server.go -o docs

echo "Swagger documentation generated successfully!"
echo "Start the server and visit http://localhost:8080/swagger/index.html to view the API documentation."