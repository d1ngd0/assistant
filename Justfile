# test runs the unit tests for the application.
test:
    go test ./...

# build first runs the tests, and then builds the binary.
build: test
    mkdir -p out/
    go build -o out/assistant ./bin

# run will build and run the application
# Source - https://stackoverflow.com/a/34672970
# Posted by Gary van der Merwe, modified by community. See post 'Timeline' for change history
# Retrieved 2026-10-19, License - CC BY-SA 3.0
run:
    #!/bin/sh
    set -e

    sigint_handler()
    {
      kill $PID
      exit
    }

    trap sigint_handler EXIT

    while true; do
      just build
      ./out/assistant &
      PID=$!
      inotifywait -e modify -e move -e create -e delete -e attrib -r `pwd`
      kill $PID
    done
