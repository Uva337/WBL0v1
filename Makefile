.PHONY: run kafka-up db-up all-up send test migrate all-down

run:
	go run ./cmd/app

kafka-up:
	docker compose up -d zookeeper kafka
	@echo "Kafka is starting, waiting 10 seconds for it to initialize..." && sleep 10

db-up:
	docker compose up -d postgres
	@echo "Postgres is starting..." && sleep 5 && echo "Postgres started."

all-up:
	docker compose up -d
	@echo "All services are starting, waiting 10 seconds for Kafka..." && sleep 10

all-down:
	docker compose down


N ?= 10
send:
	go run ./cmd/producer -n $(N)

migrate:
	go run ./cmd/migrate

test:
	go test -v ./...
