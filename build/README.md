# Container-baseimage CI/CD tool

This project use [dagger](https://github.com/dagger/dagger) as CI/CD tool to build, test and deploy images.

Please refer to the [dagger documentation](https://docs.dagger.io/) to install dagger.

# Example command lines
## Get help
```
go run main.go --help
go run main.go build --help
go run main.go test --help
go run main.go deploy --help
```

## Build and run tests
```
go mod vendor
dagger run --progress=plain go run main.go test ../Dockerfile develop --with-nonroot
```
