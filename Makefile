.PHONY: run build init

build: init
	cd ./cmd/minetest-go; go build -race

init:
	./makeinit.sh
	
run: build
	cd ./cmd/minetest-go/; ./minetest-go

runv: build
	cd ./cmd/minetest-go/; ./minetest-go -v true
