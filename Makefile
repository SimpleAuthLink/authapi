.PHONY: run demo clean

demo: clean
	@echo "Building demo image..."
	docker build -f docker/Dockerfile.demo -t demo-simpleauthlink .
	@echo "Running demo container..."
	docker run --name demo-simpleauthlink --env-file .env -p ${PORT}:80 -d demo-simpleauthlink

clean: checkport
	@echo "Cleaning up previous containers and images..."
	@docker rm -f demo-simpleauthlink 2>/dev/null || true
	@echo "Containers cleaned up"
	@docker rmi -f demo-simpleauthlink 2>/dev/null || true
	@echo "Images cleaned up"
	@echo "Cleaning up done"

checkport:
	@echo "Using default port ${PORT}"