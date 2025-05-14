$env:GOOS="linux"
$env:GOARCH="arm64"
go build -o ./bin/monitor main.go