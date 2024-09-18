css:
	npx tailwindcss -i ./assets/views/input.css -o ./assets/css/index.css

server:
	go build -C src -o ../server
	cd ../
