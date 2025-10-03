.SILENT:

POSTGRES_CONTAINER=my_postgres
POSTGRES_USER=admin
POSTGRES_PASSWORD=secret
POSTGRES_DB=mydb
POSTGRES_PORT=5432
POSTGRES_IMAGE=postgres:15

BINARY_NAME=authservice.exe

.PHONY: postgres start stop logs psql redis-cli remove build run clean

# pull Postgres
postgres:
	docker run -d \
		--name $(POSTGRES_CONTAINER) \
		-e POSTGRES_USER=$(POSTGRES_USER) \
		-e POSTGRES_PASSWORD=$(POSTGRES_PASSWORD) \
		-e POSTGRES_DB=$(POSTGRES_DB) \
		-p $(POSTGRES_PORT):5432 \
		$(POSTGRES_IMAGE)

# Start containers
start:
	docker start $(POSTGRES_CONTAINER)

# Stop containers
stop:
	docker stop $(POSTGRES_CONTAINER)

# Remove containers
remove:
	docker rm -f $(POSTGRES_CONTAINER)

# Logs
logs:
	docker logs -f $(POSTGRES_CONTAINER)

# Postgres CLI
psql:
	docker exec -it $(POSTGRES_CONTAINER) psql -U $(POSTGRES_USER) -d $(POSTGRES_DB)

# Build binary
build: 
	go build -o $(BINARY_NAME)

# Run app
run: build
	./$(BINARY_NAME)

# Clean binary
clean:
	rm -f $(BINARY_NAME)
