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
	@echo "\033[1;35m--------- Data generation ---------\033[0m"
	rm -rf send_get_scripts/sample_data
	go run generator/gen_orders.go

# --- Service --- 
run:
	@echo "\033[1;35m--------- Running service ---------\033[0m"
	@go run cmd/order_app/main.go

post-order:
	@echo "\033[1;35m--------- Posting orders ---------\033[0m"
	@./send_get_scripts/post_orders.sh
	@cat ./send_get_scripts/sample_data/uids.txt
	
check-go:
	golint ./...
	go vet ./...
	staticcheck ./...
	goimports -w .