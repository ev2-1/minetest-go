#! /bin/sh

for pl in $(ls plugins/)
do
	echo building $pl
	cd plugins/$pl
	go build -buildmode=plugin -buildvcs=false	
	cd ../..
done

rm cmd/minetest-go/plugins/*
cp -f plugins/*/*.so cmd/minetest-go/plugins/

ls cmd/minetest-go/plugins

cd cmd/minetest-go
go build .

echo "[DONE] building"
echo "<>-------------"
echo ""
