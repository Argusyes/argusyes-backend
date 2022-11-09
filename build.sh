GOOS=linux GOARCH=arm64 GOARM=7 go build -o target/linux/arm/argusyes-backend ./*.go
cp -f ./conf.toml target/linux/arm/conf.toml
GOOS=linux GOARCH=amd64 go build -o target/linux/amd/argusyes-backend ./*.go
cp -f ./conf.toml target/linux/amd/conf.toml
GOOS=darwin GOARCH=arm64 GOARM=7 go build -o target/darwin/arm/argusyes-backend ./*.go
cp -f ./conf.toml target/darwin/arm/conf.toml
GOOS=darwin GOARCH=amd64 go build -o target/darwin/amd/argusyes-backend ./*.go
cp -f ./conf.toml target/darwin/amd/conf.toml
