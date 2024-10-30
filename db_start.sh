#!/bin/sh

docker run -d \
	--rm \
	--name postgres \
    -e POSTGRES_PASSWORD=mysecretpassword \
	-e PGDATA=/var/lib/postgresql/data/pgdata \
	-v postgres_data:/var/lib/postgresql/data \
	-p 5432:5432 \
	postgres
