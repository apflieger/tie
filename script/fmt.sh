#!/usr/bin/env bash

set -e

go fmt $(go list ./... | grep -v vendor)