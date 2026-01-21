
dev:
	set -a; . ./.env; set +a; \
	go run cmd/b3/*
