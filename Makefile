.PHONY: run demo clean

demo: clean-demo
	@echo "Building demo image..."
	docker build -f docker/Dockerfile.demo -t demo-simpleauthlink .
	@echo "Running demo container..."
	docker run --name demo-simpleauthlink --env-file demo.env -p ${PORT}:80 -d demo-simpleauthlink

clean-demo:
	@echo "Cleaning up previous containers and images..."
	@docker rm -f demo-simpleauthlink 2>/dev/null || true
	@echo "Containers cleaned up"
	@docker rmi -f demo-simpleauthlink 2>/dev/null || true
	@echo "Images cleaned up"
	@echo "Cleaning up done"

api: clean-api
	@echo "Building API image..."
	docker build -f docker/Dockerfile.prod -t simpleauthlink .
	@echo "Running API container..."
	docker run --name simpleauthlink --env-file .env -p ${PORT}:80 simpleauthlink

clean-api:
	@echo "Cleaning up previous containers and images..."
	@docker rm -f simpleauthlink 2>/dev/null || true
	@echo "Containers cleaned up"
	@docker rmi -f simpleauthlink 2>/dev/null || true
	@echo "Images cleaned up"
	@echo "Cleaning up done"