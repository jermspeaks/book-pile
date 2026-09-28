.PHONY: sqlc dev test web

sqlc:
	sqlc generate

dev:
	trap 'kill 0' EXIT; go run ./cmd/server & (cd web && npm run dev); wait

test:
	go test ./...
	cd web && npm test

web:
	cd web && npm run build
