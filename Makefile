.PHONY: api run-api tidy admin-dev admin-build

tidy:
	go mod tidy

api:
	go build -o bin/chuyaji-api ./services/api/cmd/server

run-api:
	cd services/api && CHUYAJI_JWT_SECRET=dev-secret-change-me go run ./cmd/server

admin-dev:
	cd apps/admin && npm run dev

admin-build:
	cd apps/admin && npm run build
