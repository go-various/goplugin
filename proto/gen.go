//go:generate protoc -I ./ ./schema.proto ./consul.proto ./backend.proto  --go_out=plugins=grpc:.

package proto
