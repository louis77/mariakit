#!/bin/bash

# Example script showing how to use MariaKit

# Set your database connection details
# Replace these with your actual database credentials
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-3306}"
DB_USER="${DB_USER:-root}"
DB_PASSWORD="${DB_PASSWORD:-password}"
DB_NAME="${DB_NAME:-core}"

# Build the connection string
CONNECTION_STRING="${DB_USER}:${DB_PASSWORD}@tcp(${DB_HOST}:${DB_PORT})/${DB_NAME}?parseTime=true&charset=utf8mb4"

# Output directory
OUTPUT_DIR="./generated"

echo "🔍 Generating MariaDB schema code..."
echo "Database: ${DB_NAME}"
echo "Output: ${OUTPUT_DIR}"
echo ""

# Build the CLI tool if it doesn't exist
if [ ! -f "./mariakit" ]; then
    echo "Building MariaKit CLI tool..."
    go build -o mariakit ./cmd/mariakit
    echo ""
fi

# Generate all code types
./mariakit \
  -conn="${CONNECTION_STRING}" \
  -output="${OUTPUT_DIR}" \
  -config="mariakit_example.yaml"

echo ""
echo "✅ Code generation completed!"
echo "Generated files:"
echo "  - ${OUTPUT_DIR}/column_constants.go"
echo "  - ${OUTPUT_DIR}/structs.go"
echo "  - ${OUTPUT_DIR}/enum_constants.go"