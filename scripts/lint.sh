#!/bin/bash

set -e

echo "Running gofmt..."
gofmt -s -w .

echo "Running golangci-lint..."
golangci-lint run

echo "Linting completed successfully!"
