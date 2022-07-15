module github.com/ev2-1/minetest-go/cmd/minetest-go

go 1.18

require github.com/ev2-1/minetest-go v0.0.0-20220714192619-f44f43fb83cf

require (
	github.com/anon55555/mt v0.0.0-20210919124550-bcc58cb3048f // indirect
	github.com/ev2-1/minetest-go/activeobject v0.0.0-00010101000000-000000000000 // indirect
	github.com/kr/logfmt v0.0.0-20210122060352-19f9bcb100e6 // indirect
)

replace github.com/ev2-1/minetest-go => ../..

replace github.com/ev2-1/minetest-go/activeobject => ../../activeobject
