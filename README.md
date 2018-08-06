# Description

Create a simple chat with CLI interface

- Both client and server should be console apps. Users send messages to chat via stdin prompt
- Users must specify their names when joining the chat
- Each message is broadcasted to all chat members
- Same user (identified by the same name) can have multiple simultaneous clients running.
- Online status calculation: a user is "online" when he/she has at least one client running, otherwise the user is "offline".
- The server must notify all chat members when some user comes online or goes offline.

Should look like:\
server: Alice is online\
server: Bob is online\
Alice: anybody home?\
Bob: hi\
server: Alice is offline

- Client and server should be written in golang
- Any network protocols can be used
- Clean, readable code. Simplicity is a plus
(- All errors and network failures must be handled)

# Install and run

- Download from github.com

`go get -u github.com/sc-chat/test-chat`

- Go to project directory

`cd $GOPATH/src/github.com/sc-chat/test-chat`

- Run server

`go run cmd/server/main.go -a=0.0.0.0:8000 -d=true`

- Run client(s)

`go run cmd/client/main.go -a=0.0.0.0:8000 -d=true -n=Alice`

`go run cmd/client/main.go -a=0.0.0.0:8000 -d=true -n=Bob`

You can run multiple clients with the same name

`go run cmd/client/main.go -a=0.0.0.0:8000 -d=true -n=Bob`

`go run cmd/client/main.go -a=0.0.0.0:8000 -d=true -n=Bob`

For quit press Ctrl-C

# Tests

Project has small amount of unit tests

`go test ./... -v`