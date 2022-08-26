# minetest-go

a minetest server framework written in golang

working plugins / features:

- authentication without verification (`auth_nopass`)
- basic shared chat (`basic_chat`)
- basic mapblk with sqlite3 databases (`basic_map`)
- basic media announce using mth http files (`basic_media`)
	Will probably be renamed to `http_media` and `basic_media` will be "normal" mt-protocol

- nodemetas from json file (`json_nodemeta`) for documentation please wait or contact me (TODO)

# Thanks to

- Anon55555 for the AMAZING mt package

- HimbeerserverDE for being able to steal most of the networking code

- Fleckenstein for helping me make sense of minetest code

- **NOT** minetest, as i hate the technical documentation (non existant)
