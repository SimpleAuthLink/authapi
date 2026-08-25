#!/bin/bash

set -euo pipefail

SWAGGER_UI_VERSION="5.18.2"
CDN="https://cdn.jsdelivr.net/npm/swagger-ui-dist@${SWAGGER_UI_VERSION}"

mkdir -p docs

echo "Downloading Swagger UI assets..."
curl -fsSL -o docs/swagger-ui.css "$CDN/swagger-ui.css"
curl -fsSL -o docs/swagger-ui-bundle.js "$CDN/swagger-ui-bundle.js"
curl -fsSL -o docs/LICENSE.swagger-ui "$CDN/LICENSE"

echo "Generating Swagger UI entry point..."
cat > docs/index.html <<'HTML'
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>SimpleAuth.link API - Swagger UI</title>
    <link rel="stylesheet" href="./swagger-ui.css" />
    <style>
      html {
        box-sizing: border-box;
        overflow: -moz-scrollbars-vertical;
        overflow-y: scroll;
      }
      *,
      *:before,
      *:after {
        box-sizing: inherit;
      }
      body {
        margin: 0;
        background: #fafafa;
      }
    </style>
  </head>
  <body>
    <div id="swagger-ui"></div>
    <script src="./swagger-ui-bundle.js"></script>
    <script>
      window.onload = function () {
        window.ui = SwaggerUIBundle({
          url: "./swagger.yaml",
          dom_id: "#swagger-ui",
          deepLinking: true,
          presets: [SwaggerUIBundle.presets.apis],
          layout: "BaseLayout",
        });
      };
    </script>
  </body>
</html>
HTML

echo "Swagger UI assets generated successfully at docs/"