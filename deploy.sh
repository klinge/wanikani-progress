#!/bin/bash
set -e

echo "Building Go API..."
make build

echo "Building React app..."
cd web
npm run build
cd ..

echo "Build complete!"
echo "- Go binary: bin/wanikani-api"
echo "- React app: web/dist/"
echo ""
echo "Deploy with:"
echo "  scp bin/wanikani-api user@server:/path/to/api/"
echo "  scp -r web/dist/* user@server:/var/www/html/dashboard/"