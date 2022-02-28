module example


go 1.17

require mycache v0.0.0
replace mycache => ./mycache

require (
	github.com/golang/protobuf v1.5.2 // indirect
	google.golang.org/protobuf v1.27.1 // indirect
)
