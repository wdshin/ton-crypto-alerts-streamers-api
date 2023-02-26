run:
	go run .

build:
	go build .

migrate-up:
	goose -dir migrations postgres "postgres://postgres:mysecretpassword@localhost:5432/postgres?sslmode=disable" up

migrate-down:
	goose -dir migrations postgres "postgres://postgres:mysecretpassword@localhost:5432/postgres?sslmode=disable" reset
