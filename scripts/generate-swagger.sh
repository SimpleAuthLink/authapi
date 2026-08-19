#!/bin/bash

set -euo pipefail

SWAG_VERSION="v2.0.0-rc5"
SWAG="github.com/swaggo/swag/v2/cmd/swag@${SWAG_VERSION}"

echo "Downloading Go dependencies..."
go vet ./...

echo "Formatting Swagger annotations..."
go run "$SWAG" fmt -d ./

echo "Generating OpenAPI 3.1 documentation..."
mkdir -p docs
rm -f docs/swagger.yaml

go run "$SWAG" init \
    --v3.1 \
    --generalInfo api/api.go \
    --output docs \
    --outputTypes yaml \
    --parseDependency \
    --parseInternal \
    --parseDepth 4

if [ ! -f docs/swagger.yaml ]; then
    echo "Error: docs/swagger.yaml was not generated." >&2
    exit 1
fi

echo "OpenAPI documentation generated successfully at docs/openapi.yaml"