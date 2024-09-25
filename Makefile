BINARY=engine
test: clean documents generate
	go test -v -cover -covermode=atomic ./...

coverage: clean documents generate
	bash coverage.sh --html

dev: generate
	go run github.com/cosmtrek/air

run: generate
	go run .

build:
	env GOOS=linux GOARCH=amd64 go build -o ${BINARY} .

clean:
	@if [ -f ${BINARY} ] ; then rm ${BINARY} ; fi
	@find . -name *mock* -delete
	@rm -rf .cover wire_gen.go docs

docker_build:
	docker build -t boilerplate-go -f Dockerfile-local .

docker_start:
	docker-compose up --build

docker_stop:
	docker-compose down

lint-prepare:
	@echo "Installing golangci-lint" 
	curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s latest

lint:
	go run github.com/golangci/golangci-lint/cmd/golangci-lint run ./...

generate:
	go generate ./...

# migrate_create:
# 	@read -p "migration name (do not use space): " NAME \
#   	&& migrate create -ext sql -dir ./migrations/domain $${NAME}

# migrate_up:
# 	@migrate -path ./migrations/domain -database "oracle://${DB.ORACLE.WRITE.USER}:${DB.ORACLE.WRITE.PASSWORD}@${DB.ORACLE.WRITE.HOST}:${DB.ORACLE.WRITE.PORT})/${DB.ORACLE.WRITE.NAME}" up $(MIGRATION_STEP)

# migrate_down:
# 	@migrate -path ./migrations/domain -database "oracle://${DB.ORACLE.WRITE.USER}:${DB.ORACLE.WRITE.PASSWORD}@${DB.ORACLE.WRITE.HOST}:${DB.ORACLE.WRITE.PORT})/${DB.ORACLE.WRITE.NAME}" down $(MIGRATION_STEP)

# migrate_force:
# 	@read -p "please enter the migration version (the migration filename prefix): " VERSION \
#   	&& migrate -path ./migrations/domain -database "oracle://${DB.ORACLE.WRITE.USER}:${DB.ORACLE.WRITE.PASSWORD}@${DB.ORACLE.WRITE.HOST}:${DB.ORACLE.WRITE.PORT})/${DB.ORACLE.WRITE.NAME}" force $${VERSION}

# migrate_version:
# 	@migrate -path ./migrations/domain -database "oracle://${DB.ORACLE.WRITE.USER}:${DB.ORACLE.WRITE.PASSWORD}@${DB.ORACLE.WRITE.HOST}:${DB.ORACLE.WRITE.PORT})/${DB.ORACLE.WRITE.NAME}" version 

# migrate_drop:
# 	@migrate -path ./migrations/domain -database "oracle://${DB.ORACLE.WRITE.USER}:${DB.ORACLE.WRITE.PASSWORD}@${DB.ORACLE.WRITE.HOST}:${DB.ORACLE.WRITE.PORT})/${DB.ORACLE.WRITE.NAME}" drop
	
.PHONY: test coverage engine clean build docker run stop lint-prepare lint documents generate

deploy:
	scp -P 3157 ./engine quadran@173.249.36.204:/home/quadran/
	# scp -P 3157 ./.env quadran@173.249.36.204:/home/quadran/
	scp -P 3157 ./docs/* quadran@173.249.36.204:/home/quadran/docs