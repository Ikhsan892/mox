
.PHONY: up down restart logs simulate simulate-ps drain clean

# Start environment
up:
	docker compose up -d

# Stop environment
down:
	docker compose down

# Restart environment (reset state)
restart: down up

# Show logs
logs:
	docker compose logs -f

# Simulate Load (Bash/Linux/WSL version)
simulate:
	@echo "Running 100 requests..."
	@for i in $$(seq 1 100); do \
		served=$$(curl -s http://localhost:1111 | grep "PID"); \
		echo "[$$i] Handled by: $$served"; \
	done

# Simulate Load (PowerShell version - call existing script)
simulate-ps:
	powershell ./simulate_load.ps1

# Drain haproxy-1 (requires separate terminal or background)
drain:
	@echo "Draining haproxy-1..."
	docker exec quickstart-haproxy-1-1 bash -c "apt-get update >/dev/null && apt-get install -y socat >/dev/null && echo 'set maxconn frontend main 0' | socat stdio UNIX-CONNECT:/var/lib/haproxy/sockets/stats-haproxy-1.sock"
	@echo "Drain command sent. Check simulation output for traffic shift."

# Clean up anything left
clean:
	docker compose down -v
