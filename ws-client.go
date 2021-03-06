package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"gopkg.in/readline.v1"
)

var (
	VERSION = "// filled in by build //"

	showVersion = flag.Bool("version", false, "show version number and exit")
	verbose     = flag.Bool("v", false, "verbose output")
	useTsPrefix = flag.Bool("ts-prefix", true, "timestamp prefix")
)

func main() {
	flag.Parse()

	if *showVersion || *verbose || len(os.Args) < 2 {
		fmt.Printf("ws-client %s \n", VERSION)
		fmt.Println("https://github.com/oliver006/ws-client/")
		fmt.Println()
		if *showVersion {
			os.Exit(0)
		}
	}

	if len(os.Args) < 2 {
		fmt.Println("Usage:")
		fmt.Println("  ws-client <<ws:// or wss:// URL>>")
		os.Exit(-1)
	}

	addr := os.Args[len(os.Args)-1]
	if !strings.HasPrefix(addr, "ws://") && !strings.HasPrefix(addr, "wss://") {
		addr = "ws://" + addr
	}

	if *verbose {
		fmt.Printf("connecting to %s\n", addr)
	}
	c, _, err := websocket.DefaultDialer.Dial(addr, nil)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	defer c.Close()
	fmt.Printf("connected to  %s\n", addr)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	rl, err := readline.NewEx(&readline.Config{
		Prompt: "\033[32m»\033[0m ",
	})
	if err != nil {
		panic(err)
	}

	defer rl.Close()

	if *useTsPrefix {
		tsStr := time.Now().Format("[15:04] ")
		rl.SetPrompt(fmt.Sprintf("%s\033[32m»\033[0m ", tsStr))
		go func() {
			c := time.Tick(30 * time.Second)
			for now := range c {
				tsStr := now.Format("[15:04] ")
				rl.SetPrompt(fmt.Sprintf("%s\033[32m»\033[0m ", tsStr))
				rl.Refresh()
			}
		}()
	}
	go func() {
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				fmt.Printf("<<server: %s>>\n", err)
				close(interrupt)
				return
			}
			tsStr := ""

			if *useTsPrefix {
				tsStr = time.Now().Format("[15:04] ")
			}
			io.WriteString(rl.Stdout(), fmt.Sprintf("%s\033[31m«\033[0m %s\n", tsStr, message))
		}
	}()

	go func() {
		for {
			line, err := rl.Readline()
			if err == readline.ErrInterrupt {
				interrupt <- os.Interrupt
				return
			} else if err != nil {
				close(interrupt)
				return
			}
			if len(line) == 0 {
				continue
			}
			err = c.WriteMessage(websocket.TextMessage, []byte(line))
			if err != nil {
				fmt.Println("err:", err)
				return
			}
		}
	}()

	select {
	case <-interrupt:
		c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		fmt.Println("<<client: sent websocket close frame>>")
		c.Close()
		os.Exit(0)
	}
}
