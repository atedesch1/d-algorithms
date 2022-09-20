all: process shared_resource

process: cmd/process/main.go
	go build -o bin/process cmd/process/main.go

shared_resource: cmd/shared_resource/main.go
	go build -o bin/shared_resource cmd/shared_resource/main.go


clean:
	rm -rf bin/*