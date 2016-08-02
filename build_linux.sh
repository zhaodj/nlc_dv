#/usr/local/bin/fish
env GOARCH=amd64 GOOS=linux CGO_ENABLED=0 go build app.go
