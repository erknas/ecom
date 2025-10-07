SERVICES := user-service
	
#user-service
build-user-service:
	@cd services/user-service && go build -o ../../bin/user-service cmd/main.go && cd ../..
run-user-service-local: build-user-service
	@./bin/user-service -config=services/user-service/config/local.yaml -env=services/user-service/.env.local
test-user-service:
	@go clean -testcache
	@cd services/user-service && go test -cover -v ./...
migrate-user-service:
	@go run tools/migrator/main.go -path=migrations/users -env=services/user-service/.env.local