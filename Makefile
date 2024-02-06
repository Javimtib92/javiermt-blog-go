tailwind-build:
	./tailwindcss -i ./web/styles.css -o ./web/static/css/styles.css --minify

dev:
	go run main.go

prod:
	go run main.go -https


.PHONY: tailwind-build dev