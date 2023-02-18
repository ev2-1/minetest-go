# minetest-go

a minetest server framework thing written in golang

working plugins / features:

- per-account storage (`minetest.ClientData`)
- inventories (current_player and detached) (`inventory`)
- authentication without verification (`auth_nopass`)
- basic shared chat (`basic_chat`)
- map
  - compatability with minetest maps (`map/minetest_mapdriver`)
    - using (`github.com/eliasFleckenstein03/mtmap`)
  - multible dimensions (4d, extra uint16 value)
- texture media announce using mth http files (`http_media`)
- mesh media announce using the standard mt method (`mt_media`)
- nodemetas from json file (`json_nodemeta`)
  - can be generated using [meta_dumper](//github.com/ev2-1/meta_dumper)

# Thanks to

- Anon55555 for the AMAZING mt package

- HimbeerserverDE for being able to steal most of the networking code from

- Fleckenstein for helping me make sense of minetest code

- **NOT** minetest, as the technical documentation is non existant
