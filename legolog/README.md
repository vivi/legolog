## How to run tests
Run `make proto` from the project directory to compile the protobufs. Then open two terminals.

Terminal 1
1. cd into legolog
2. run the server using `go run main.go` 


Terminal 2
1. cd into `legolog/client`
2. run the test client (client_test.go) using `go test`
