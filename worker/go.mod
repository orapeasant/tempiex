module github.com/tempiex/worker

go 1.25

require (
	github.com/goccy/go-yaml v1.15.23
	github.com/rs/zerolog v1.34.0
	github.com/tempiex/tempiex v0.0.0
	google.golang.org/grpc v1.72.2
)

require (
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	golang.org/x/net v0.40.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.25.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250519155744-55703ea1f237 // indirect
	google.golang.org/protobuf v1.36.6 // indirect
)

replace github.com/tempiex/tempiex => ../tempiex
