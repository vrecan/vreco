#!/bin/bash

go build && go test -v ./... && npx tailwindcss -i ./src/tailwindcss/input.css -o ./static/css/mystyles.css && ./vreco
