# ws-client
A simple command line websocket client, written in Golang


## Installation

```
$ go install github.com/oliver006/ws-client
```


## Usage

```
$ ws-client ws://echo.websocket.org
connected to  ws://echo.websocket.org
» yo
« yo
» sup?
« sup?
» ^C
<<client: sent websocket close frame>>
$
```

