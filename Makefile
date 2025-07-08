run:
	@echo "Starting Go backend and React frontend..."	
	@go run -race cmd/main.go & (cd web && npm run dev -- --host 0.0.0.0)
	@air
