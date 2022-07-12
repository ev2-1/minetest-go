#! /bin/sh

rm cmd/minetest-go/plugins/*.so

for pl in $(ls plugins/)
do
	echo building $pl
	cd plugins/$pl
	go get github.com/ev2-1/minetest-go
	go build -buildmode=plugin -buildvcs=false	
	cd ../..

	cp plugins/*/*.so cmd/minetest-go/plugins/
done

ls cmd/minetest-go/plugins

cd cmd/minetest-go
go get github.com/ev2-1/minetest-go
go build .
