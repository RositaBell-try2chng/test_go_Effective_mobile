.PHONY: all build down clean re rebuild logs deps

all:
	docker-compose up -d

build:
	docker-compose up --build

down:
	docker-compose down

clean:
	docker-compose down -v --rmi all
	rm -rf bin/

re: clean all

rebuild: clean build

logs:
	docker-compose logs -f

deps:
	go mod download
	go mod tidy
