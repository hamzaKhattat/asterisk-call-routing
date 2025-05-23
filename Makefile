.PHONY: build run test clean migrate

build:
   go build -o bin/router cmd/router/main.go

run: build
   ./bin/router -config configs/config.json

interactive: build
   ./bin/router -i

migrate:
   mysql -u root -ptemppass < migrations/001_initial_schema.sql

import-dids: build
   ./bin/router -import-dids $(FILE)

stats: build
   ./bin/router -stats

cleanup: build
   ./bin/router -cleanup

test:
   go test ./...

clean:
   rm -rf bin/

docker-build:
   docker build -t asterisk-router:latest .

docker-run:
   docker run -d \
   	--name asterisk-router \
   	-p 8000:8000 \
   	-e DB_HOST=mysql \
   	-e DB_PASSWORD=temppass \
   	--link mysql:mysql \
   	asterisk-router:latest
