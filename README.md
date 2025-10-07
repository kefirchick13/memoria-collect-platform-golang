// Memoria

// docker run --name Memoria -P -p 127.0.0.1:5433:5432 -e POSTGRES_PASSWORD="1234" postgres:alpine

// migrate -path ./schema -database 'postgres://postgres:1234@127.0.0.1:5433/postgres?sslmode=disable' up