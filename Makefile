.PHONY: demo api swagger-ui clean-demo clean-api

demo: clean-demo
	@echo "Building demo image..."
	docker build -f docker/Dockerfile.demo -t demo-simpleauthlink .
	@echo "Running demo container..."
	@set -a; . ./.env; set +a; docker run --name demo-simpleauthlink --env-file .env -p $${PORT}:$${PORT} demo-simpleauthlink

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
	@set -a; . ./.env; set +a; docker run --name simpleauthlink --env-file .env -p $${PORT}:$${PORT} simpleauthlink

clean-api:
	@echo "Cleaning up previous containers and images..."
	@docker rm -f simpleauthlink 2>/dev/null || true
	@echo "Containers cleaned up"
	@docker rmi -f simpleauthlink 2>/dev/null || true
	@echo "Images cleaned up"
	@echo "Cleaning up done"

swagger-ui:
	./scripts/generate-swagger.sh
	@trap 'git checkout HEAD -- docs/swagger.yaml' EXIT; \
	docker run --rm \
		-p 8081:8080 \
		-e SWAGGER_JSON=/spec/swagger.yaml \
		-v "$$PWD/docs:/spec:ro" \
		docker.swagger.io/swaggerapi/swagger-ui