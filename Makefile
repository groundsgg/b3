# SPDX-License-Identifier: AGPL-3.0-or-later

dev:
	set -a; . ./.env; set +a; \
	go run cmd/b3/*
