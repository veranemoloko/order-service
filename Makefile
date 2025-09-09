.PHONY: up down gen run check-go rebuild

MAKEFLAGS += --no-print-directory

# --- Docker ---
up:
	docker compose up -d

down:
	docker compose down -v

rebuild: down up

# --- Data generation ---
gen:
	rm -rf send_get_scripts/sample_data
	go run generator/gen_orders.go

# --- Service --- 
run:
	@echo "\033[1;35m--------- Running service ---------\033[0m"
	@go run cmd/order_app/main.go

post-get-order:
	@echo "\033[1;35m--------- Generating orders ---------\033[0m"
	@$(MAKE) gen
	@sleep 2
	@echo "\033[1;35m--------- Posting orders ---------\033[0m"
	@./send_get_scripts/post_orders.sh
	@sleep 5
	@echo "\033[1;35m--------- Getting orders ---------\033[0m"
	@./send_get_scripts/get_orders.sh
	
check-go:
	golint ./...
	go vet ./...
	staticcheck ./...
	goimports -w .