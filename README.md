# minetest-go

a minetest server framework thing written in golang

working plugins / features:

- per-account storage (`minetest.ClientData`)
- inventories (current_player and detached) (`inventory`)
- authentication without verification (`auth_nopass`)
- basic shared chat (`basic_chat`)
- basic mapblk with sqlite3 databases (`basic_map`)
  - supports minetest map format though (`github.com/eliasFleckenstein03/mtmap`)
- basic media announce using mth http files (`basic_media`)
	Will probably be renamed to `http_media` and `basic_media` will be "normal" mt-protocol
- nodemetas from json file (`json_nodemeta`)
  - can be generated using [meta_dumper](//github.com/ev2-1/meta_dumper)

# Thanks to

- Anon55555 for the AMAZING mt package

- HimbeerserverDE for being able to steal most of the networking code from

- Fleckenstein for helping me make sense of minetest code

- **NOT** minetest, as the technical documentation is non existant
