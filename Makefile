.PHONY: build up down logs db

build:
docker build -t library-api ./

up:
docker-compose up -d --build

down:
docker-compose down

logs:
docker-compose logs -f

db:
docker exec -it library-mysql mysql -u$$DB_USER -p$$DB_PASSWORD