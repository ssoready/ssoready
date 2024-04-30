.PHONY: queries
queries:
	pg_dump --schema-only 'postgres://postgres:password@localhost?sslmode=disable' > sqlc/schema.sql
	sqlc -f ./sqlc/sqlc.yaml generate

.PHONY: proto
proto:
	npx buf generate internal/proto
