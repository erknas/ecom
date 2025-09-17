SERVICES := user-service

build:
	@for service in ${SERVICES}; do \
		cd services/$$service && go build -o ../../bin/$$service cmd/main.go && cd ../..; \
	done

run: build
	@./bin/user-service -config=services/user-service/config/config.yaml

build-user-service:
	@cd services/user-service && go build -o ../../bin/user-service cmd/main.go && cd ../..

run-user-service: build-user-service
	@./bin/user-service -config=services/user-service/config/config.yaml

migrate-user-service:
	@go run tools/migrator/main.go -path=migrations/users