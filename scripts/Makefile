.PHONY: postgres postgres-stop postgres-rm

postgres:
	docker run --name postgres-db \
	-e POSTGRES_USER=postgres \
	-e POSTGRES_PASSWORD=postgres \
	-e POSTGRES_DB=blockscan \
	-p 5432:5432 \
	-v pgdata:/var/lib/postgresql/data \
	-d postgres:16

postgres-stop:
	docker stop postgres-db

postgres-rm:
	docker rm postgres-db