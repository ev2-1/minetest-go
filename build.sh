#! /bin/sh

for pl in $(ls minetest-go/plugins/)
do
	echo building $pl
	cd minetest-go/plugins/$pl
	go get github.com/ev2-1/minetest-go
	go build -buildmode=plugin -buildvcs=false	
	cd ../../..

	cp minetest-go/plugins/*/*.so minetest-go/cmd/minetest-go/plugins/
done

ls minetest-go/cmd/minetest-go/plugins

cd minetest-go/cmd/minetest-go
go get github.com/ev2-1/minetest-go
go build .
