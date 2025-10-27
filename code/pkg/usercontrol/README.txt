There are two pieces of Go source here, one for Linux and one for Windows.

By default, this compiles for Windows.

To compile for Linux:

export GOOS=Linux
go build -tags linux