#!/bin/sh

THIS_FILE_DIR="$(cd "$(dirname "$0")" && pwd)"

PROJECT_DIR="$(cd "${THIS_FILE_DIR}/.." && pwd)"

go build

