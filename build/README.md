# CI/CD tool

## Get Help
```
go run main.go --help
```

## Run Tests
```
go mod vendor
go run main.go job test ../Dockerfile --target base --export ../bin
go run main.go job test ../Dockerfile --target nonroot --tags-suffix "-nonroot"
```
