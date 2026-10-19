# test runs the unit tests for the application.
test:
    go test ./...

# build first runs the tests, and then builds the binary.
build: test
    mkdir -p out/
    go build -o out/assistant ./...
