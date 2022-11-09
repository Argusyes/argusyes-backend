GOOS=linux GOARCH=arm64 GOARM=7 go build -o target/linux/arm/argusyes-backend ./*.go
GOOS=linux GOARCH=amd64 go build -o target/linux/amd/argusyes-backend ./*.go
GOOS=darwin GOARCH=arm64 GOARM=7 go build -o target/darwin/arm/argusyes-backend ./*.go
GOOS=darwin GOARCH=amd64 go build -o target/darwin/amd/argusyes-backend ./*.go