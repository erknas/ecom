SERVICES := user-service

proto:
	@for service in ${SERVICES}; \
	do \
		cd services/$$service && \
			protoc --go_out=. --go_opt=paths=source_relative \
			--go-grpc_out=. --go-grpc_opt=paths=source_relative \
			proto/*.proto && cd ../..; \
	done

build:
	@for service in ${SERVICES}; do \
		cd services/$$service && go build -o ../../bin/$$service cmd/main.go && cd ../..; \
	done

run: build
	@./bin/user-service -config=services/user-service/config/config.yaml

#user-service
build-user-service:
	@cd services/user-service && go build -o ../../bin/user-service cmd/main.go && cd ../..
run-user-service-local: build-user-service
	@./bin/user-service -config=services/user-service/config/local.yaml -env=services/user-service/.env.local
migrate-user-service:
	@go run tools/migrator/main.go -path=migrations/users -env=services/user-service/.env.local
	
PHONY: proto