package main

//main --port_in=8090 --port_out=8091 --name=user1
import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"github.com/gizak/termui"
)

var portIn string
var portOut string
var name string
var strs []string
var mutex = &sync.Mutex{}

func main() {

	initParams()
	go listening()

	err := termui.Init()
	if err != nil {
		panic(err)
	}
	defer termui.Close()

	list := termui.NewList()
	list.Items = strs
	list.ItemFgColor = termui.ColorYellow
	list.BorderLabel = "Chat"
	list.Height = 20
	list.Y = 0

	inputText := ""
	input := termui.NewPar(inputText)
	input.Height = 3
	input.TextFgColor = termui.ColorWhite
	input.BorderLabel = "Message"
	input.BorderFg = termui.ColorCyan

	draw := func() {
		list.Items = strs
		input.Text = inputText
		termui.Render(list, input)
	}

	termui.Body.AddRows(
		termui.NewRow(termui.NewCol(11, 0, list)),
		termui.NewRow(termui.NewCol(11, 0, input)))

	termui.Body.Align()

	termui.Render(termui.Body)
	termui.Handle("/sys/wnd/resize", func(e termui.Event) {
		termui.Body.Width = termui.TermWidth()
		termui.Body.Align()
		termui.Clear()
		termui.Render(termui.Body)
	})
	termui.Handle("/timer/1s", func(termui.Event) {
		draw()
	})
	termui.Handle("/sys/kbd/C-c", func(termui.Event) {
		termui.StopLoop()
	})
	termui.Handle("/sys/kbd", func(e termui.Event) {
		key := e.Path[9:]

		switch key {
		case "<backspace>":
			l := len(inputText)
			if l > 0 {
				inputText = inputText[0 : l-1]
			}

		case "<space>":
			inputText += " "
		case "<enter>":
			send(name + ": " + inputText)
			inputText = ""
		default:
			if len(key) == 1 {
				inputText += key
			}
		}

		draw()

	})
	termui.Loop()
}

func initParams() {
	if len(os.Args) != 4 {
		log.Printf("Ошибка количества параметров!")
		os.Exit(1)
	}

	flag.StringVar(&portIn, "port_in", "8080", "Port in")
	flag.StringVar(&portOut, "port_out", "8081", "Port out")
	flag.StringVar(&name, "name", "name", "Name")
	flag.Parse()

	fmt.Println(portIn)
	fmt.Println(portOut)
	fmt.Println(name)

}

func addLisetRow(msg string) {
	t := time.Now().Format(time.Stamp)
	mutex.Lock()
	strs = append(strs, t+" "+" "+msg)

	if len(strs) > 18 {
		strs = strs[1:]
	}
	mutex.Unlock()
}

func send(msg string) {
	addLisetRow(msg)

	conn, err := net.Dial("tcp", "127.0.0.1:"+portOut)
	if err != nil {
		panic(err)
	}

	defer conn.Close()

	conn.Write([]byte(msg))
}

func listening() {
	listener, err1 := net.Listen("tcp", ":"+portIn)

	if err1 != nil {
		panic(err1)
	}
	defer listener.Close()

	for {
		conn, err2 := listener.Accept()
		if err2 != nil {
			panic(err2)
		}
		defer conn.Close()

		message := make([]byte, 1024)

		_, err3 := conn.Read(message[:]) // recv data
		if err3 != nil {
			continue
		}

		addLisetRow(string(message))
	}
}
