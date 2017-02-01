#!/usr/bin/env bash

set -e

go test -cover $(go list ./... | grep -v vendor)