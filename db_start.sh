#!/bin/sh

docker run -d \
	--rm \
	--name postgres \
    -e TZ=Europe/Vienna \
	-e PGDATA=/var/lib/postgresql/data/pgdata \
	-v postgres_data:/var/lib/postgresql/data \
	-p 5432:5432 \
	postgres
