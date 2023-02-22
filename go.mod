module github.com/ev2-1/minetest-go

go 1.19

require (
	github.com/EliasFleckenstein03/mtmap v0.4.0
	github.com/anon55555/mt v0.0.0-20210919124550-bcc58cb3048f
	github.com/mattn/go-sqlite3 v1.14.15
)

require github.com/g3n/engine v0.2.0

require github.com/mattn/go-shellwords v1.0.12

require (
	github.com/go-yaml/yaml v2.1.0+incompatible
	github.com/kevinburke/nacl v0.0.0-20210405173606-cd9060f5f776
	github.com/spf13/pflag v1.0.5
)

require (
	github.com/kr/pretty v0.2.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
)

replace debug_sync => ../debug_sync

replace github.com/ev2-1/mineclone-go => ../mineclone-go
