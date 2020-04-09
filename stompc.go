package main
import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/go-stomp/stomp"
)

var serverAddr = flag.String("server", "localhost:61613", "STOMP server endpoint")
var name = flag.String("name", "Guest", "Chat name")
var stop = make(chan bool)
var queueName = "/queue/client_test"
var subscribed = make(chan bool)

var options []func(*stomp.Conn) error = []func(*stomp.Conn) error{
	stomp.ConnOpt.Login("guest", "guest"),
	stomp.ConnOpt.Host("/"),
}

func recvMsgs() {
	defer func() {
		stop <- true
	}()

	conn, err := stomp.Dial("tcp", *serverAddr, options...)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	sub, err := conn.Subscribe(queueName, stomp.AckAuto)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	close(subscribed)

	for {
		msg := <-sub.C
		text := string(msg.Body)
		if strings.Split(text, ":")[0] != *name {
			fmt.Println(text)
		}
	}

}

func sendMsgs() {
	defer func() {
		stop <- true
	}()

	conn, err := stomp.Dial("tcp", *serverAddr, options...)
	if err != nil {
		println("cannot connect to server", err.Error())
		return
	}
	in := bufio.NewReader(os.Stdin)

	for {
		line, err := in.ReadString('\n')
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		line = *name + ": " + strings.TrimRight(line, "\n")
		err = conn.Send(queueName, "text/plain", []byte(line), nil)
	}
}

func main() {
	flag.Parse()

	go recvMsgs()

	<-subscribed

	go sendMsgs()

	<-stop
	<-stop
}
