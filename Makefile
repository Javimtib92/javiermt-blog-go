tailwind-build:
	./tailwindcss -i ./web/styles.css -o ./web/static/css/styles.css --minify

dev:
	go run main.go

prod:
	go run main.go -https

docker-build:
	docker build -t main .

docker-run:
	docker run --rm -p 8080:8080 -it main

.PHONY: tailwind-build dev prod docker-build docker-run