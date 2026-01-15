# run with basic auth: admin:admin (admin user) and vistor:vistor (viewer user)
dev:
	set -a; . ./.env; set +a; \
	go run cmd/b3/*

dev-basic:
	set -a; . ./.env.basic_auth.example; set +a; \
	go run cmd/b3/*