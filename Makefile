run:
	go run .

build:
	go build .

migrate-up:
	goose -dir migrations postgres "postgres://postgres:mysecretpassword@localhost:5432/postgres?sslmode=disable" up

migrate-down:
	goose -dir migrations postgres "postgres://postgres:mysecretpassword@localhost:5432/postgres?sslmode=disable" reset

migrate-prod-up:
	goose -dir migrations postgres "postgres://donatedb_user:82OUnHooVqrAzZmiXXjZ6rua6Vevm9Qm@dpg-cfts93qrrk0c837gcejg-a.oregon-postgres.render.com/donatedb" up
