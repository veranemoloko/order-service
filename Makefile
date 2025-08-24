.PHONY: up down gen topic-create topic-list check-data run check-go rebuild clean

MAKEFLAGS += --no-print-directory

# --- Docker ---
up:
	docker-compose up -d

down:
	docker-compose down -v

# --- Data generation ---
gen:
	go run generator/gen_orders.go

# --- Kafka ---
topic-create:
	docker-compose exec -T kafka \
		kafka-topics --create --topic orders \
		--partitions 2 --replication-factor 1 \
		--bootstrap-server kafka:9092 || true

topic-list:
	docker-compose exec -T kafka \
		kafka-topics --list --bootstrap-server kafka:9092

check-data:
	docker exec -it kafka kafka-console-consumer \
		--bootstrap-server localhost:9092 \
		--topic orders \
		--from-beginning

run:
	@echo "\033[1;35m--------- Creating Kafka topic 'orders' ---------\033[0m"
	-@$(MAKE) topic-create
	@sleep 3
	@echo "\033[1;35m--------- Running service ---------\033[0m"
	@go run cmd/service/main.go

post-get-order:
	@$(MAKE) clean
	@echo "\033[1;35m--------- Generating orders ---------\033[0m"
	@$(MAKE) gen
	@sleep 3
	@echo "\033[1;35m--------- Posting orders ---------\033[0m"
	@./send_get_scripts/post_orders.sh
	@sleep 10
	@echo "\033[1;35m--------- Getting orders ---------\033[0m"
	@./send_get_scripts/get_orders.sh
	

clean:
	rm -rf send_get_scripts/sample_data

rebuild: clean down run

check-go:
	golint ./...
	go vet ./...
	staticcheck ./...
	goimports -w .