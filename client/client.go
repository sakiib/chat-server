package client

import (
	"bufio"
	"fmt"
	"github.com/sakiib/chat-server/utils"
	"net"
	"os"
	"sync"
)

var wg sync.WaitGroup

func Read(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		str, err := reader.ReadString('\n')
		if err != nil {
			wg.Done()
			return
		}
		fmt.Print(str)
	}
}

func Write(conn net.Conn) {
	reader := bufio.NewReader(os.Stdin)
	writer := bufio.NewWriter(conn)

	for {
		str, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		_, err = writer.WriteString(str)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		err = writer.Flush()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}

func ConnectClient() {
	wg.Add(2)

	conn, err := net.Dial(utils.TYPE, utils.PORT)
	if err != nil {
		fmt.Println(err)
	}

	go Read(conn)
	go Write(conn)

	wg.Wait()
}
