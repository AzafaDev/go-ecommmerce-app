include .env

migrate-up:
	migrate -path migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path migrations -database "$(DATABASE_URL)" down 1

migrate-force:
	migrate -path migrations -database "$(DATABASE_URL)" force $(VERSION)

seed:
	go run ./cmd/seed