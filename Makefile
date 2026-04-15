.PHONY: api run-api tidy

tidy:
	go mod tidy

api:
	go build -o bin/chuyaji-api ./services/api/cmd/server

run-api:
	cd services/api && CHUYAJI_JWT_SECRET=dev-secret-change-me go run ./cmd/server
