#!/bin/bash

set -euo pipefail

SWAGGER_UI_VERSION="5.18.2"
CDN="https://cdn.jsdelivr.net/npm/swagger-ui-dist@${SWAGGER_UI_VERSION}"

mkdir -p docs/api

echo "Downloading Swagger UI assets..."
curl -fsSL -o docs/api/swagger-ui.css "$CDN/swagger-ui.css"
curl -fsSL -o docs/api/swagger-ui-bundle.js "$CDN/swagger-ui-bundle.js"
curl -fsSL -o docs/api/LICENSE.swagger-ui "$CDN/LICENSE"

echo "Generating Swagger UI entry point..."
cat > docs/api/index.html <<'HTML'
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
      // Swagger UI updates the URL hash for deep links via history.pushState,
      // which pollutes the browser history (Back requires several clicks to
      // leave the page). Downgrade to replaceState so this page never adds
      // history entries: Back returns to the docs in one click, and deep
      // links (e.g. #/apps/post_apps) still open the right operation.
      (function () {
        var pushState = window.history.pushState.bind(window.history);
        window.history.pushState = function (state, title, url) {
          return window.history.replaceState(state, title, url);
        };
      })();

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

echo "Swagger UI assets generated successfully at docs/api/"