tailwind-build:
	./tailwindcss -i ./web/styles.css -o ./web/static/css/styles.css --minify

dev:
	go run main.go


.PHONY: tailwind-build dev