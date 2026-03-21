POSTGRES_CONTAINER = go-project-278-postgres
TEST_POSTGRES_CONTAINER = go-project-278-postgres-test
POSTGRES_IMAGE = postgres:16

test:
	go test --race ./...

lint:
	golangci-lint run ./...

run:
	go run main.go

start-frontend:
	npx start-hexlet-url-shortener-frontend

db-up:
	docker run --name $(POSTGRES_CONTAINER) \
		-e POSTGRES_PASSWORD=postgres \
		-e POSTGRES_DB=link_shortener \
		-p 5432:5432 \
		-d $(POSTGRES_IMAGE)

db-down:
	docker stop $(POSTGRES_CONTAINER)
	docker rm $(POSTGRES_CONTAINER)

test-db-up:
	docker run --name $(TEST_POSTGRES_CONTAINER) \
		-e POSTGRES_PASSWORD=postgres \
		-e POSTGRES_DB=link_shortener_test \
		-p 5432:5432 \
		-d $(POSTGRES_IMAGE)

test-db-down:
	docker stop $(TEST_POSTGRES_CONTAINER)
	docker rm $(TEST_POSTGRES_CONTAINER)
