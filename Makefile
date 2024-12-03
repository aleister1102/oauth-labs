.PHONY: all
all: build

.PHONY: build
build:
	cd ./configure && make || true && cd ..
	cd ./labindex && make || true && cd ..
	cd ./victim && make || true && cd ..
	cd ./lab01/server && make || true && cd ..
	cd ./lab01/client && make || true && cd ..
	cd ./lab02/server && make || true && cd ..
	cd ./lab02/client && make || true && cd ..
	cd ./lab03/server && make || true && cd ..
	cd ./lab03/client && make || true && cd ..
	cd ./lab04/server && make || true && cd ..
	cd ./lab04/client && make || true && cd ..
	cd ./lab05/server && make || true && cd ..
	cd ./lab05/client && make || true && cd ..

.PHONY: lint
lint:
	cd ./oalib && make lint || true && cd ..
	cd ./labindex && make lint || true && cd ..
	cd ./victim && make lint || true && cd ..
	cd ./lab01/server && make lint || true && cd ..
	cd ./lab01/client && make lint || true && cd ..
	cd ./lab02/server && make lint || true && cd ..
	cd ./lab02/client && make lint || true && cd ..
	cd ./lab03/server && make lint || true && cd ..
	cd ./lab03/client && make lint || true && cd ..
	cd ./lab04/server && make lint || true && cd ..
	cd ./lab04/client && make lint || true && cd ..
	cd ./lab05/server && make lint || true && cd ..
	cd ./lab05/client && make lint || true && cd ..

.PHONY: docker
docker:
	docker compose -f ./docker-compose.yaml build --parallel

.PHONY: labs
labs:
	docker compose -f ./docker-compose.yaml up -d

.PHONY: lab01
lab01:
	docker compose -f ./docker-compose.yaml up -d server-01 client-01

.PHONY: lab02
lab02:
	docker compose -f ./docker-compose.yaml up -d server-02 client-02

.PHONY: lab03
lab03:
	docker compose -f ./docker-compose.yaml up -d server-03 client-03

.PHONY: lab04
lab04:
	docker compose -f ./docker-compose.yaml up -d server-04 client-04

.PHONY: lab05
lab05:
	docker compose -f ./docker-compose.yaml up -d server-05 client-05

.PHONY: labsdown
labsdown:
	docker compose -f ./docker-compose.yaml down -v

.PHONY: devup
devup:
	docker compose -f ./docker-compose.dev.yaml up -d

.PHONY: devdown
devdown:
	docker compose -f ./docker-compose.dev.yaml down -v

.PHONY: config
config:
	cd ./configure && make && cd ..
	./configure/bin/configure

.PHONY: configure
configure:
	cd ./configure && make && cd ..
	./configure/bin/configure
