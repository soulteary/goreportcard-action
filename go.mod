module github.com/soulteary/goreportcard-action

go 1.27.4

require (
	github.com/client9/misspell v0.3.4 // indirect
	github.com/fzipp/gocyclo v0.6.0 // indirect
	github.com/gordonklaus/ineffassign v0.2.0 // indirect
	golang.org/x/mod v0.43.0 // indirect
	golang.org/x/sync v0.25.0 // indirect
	golang.org/x/tools v0.53.0 // indirect
)

tool (
	github.com/client9/misspell/cmd/misspell
	github.com/fzipp/gocyclo/cmd/gocyclo
	github.com/gordonklaus/ineffassign
)
