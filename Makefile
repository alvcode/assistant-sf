help:
	go run cmd/main.go --help

init:
	go run cmd/main.go init

auth:
	go run cmd/main.go auth

from-disk:
	go run cmd/main.go from-disk

to-disk:
	go run cmd/main.go to-disk

sync-server:
	go run cmd/main.go sync --head server

sync-local:
	go run cmd/main.go sync --head local