.PHONY: run demo clean

demo: clean
	@echo "Building demo image..."
	docker build -f docker/Dockerfile.demo -t demo-simpleauthlink .
	@echo "Running demo container..."
	docker run --name demo-simpleauthlink --env-file .env -p 8080:8080 -d demo-simpleauthlink

clean:
	@echo "Cleaning up previous containers and images..."
	-docker rm -f demo-simpleauthlink 2>/dev/null || true
	-docker rmi -f demo-simpleauthlink 2>/dev/null || true
