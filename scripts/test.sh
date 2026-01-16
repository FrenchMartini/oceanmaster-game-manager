#!/bin/bash

set -e

echo "Running tests..."
go test ./...

echo "Tests completed successfully!"
