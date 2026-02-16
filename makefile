build:
	@go build -o tuigote .

run: build
	@./tuigote
