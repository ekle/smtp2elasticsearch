package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
)

const (
	HOST = "localhost"
	PORT = "25"
)

func main() {
	l, err := net.Listen("tcp", HOST+":"+PORT)
	if err != nil {
		fmt.Println("net.Listen:", err.Error())
		os.Exit(1)
	}
	defer l.Close()
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Accept", err.Error())
			return
		}
		go handle(conn)
	}
}

type message struct {
	Header string
	Body   string
}

func handle(conn net.Conn) {
	defer conn.Close()
	defer fmt.Println("connection closed")
	fmt.Println("---------------------")
	buf := make([]byte, 2048)
	conn.Write([]byte("220 whatever\r\n"))
	DATA := []byte{}
	HEADER := []byte{}
	for true {
		size, err := conn.Read(buf)
		if err != nil {
			break
		}
		line := buf[:size]
		CMD := string(line[0:4])
		fmt.Println("CMD: " + CMD)
		switch CMD {
		case "HELO":
			conn.Write([]byte("250 whatever\r\n"))
		default:
			conn.Write([]byte("250 OK\r\n"))
			HEADER = append(HEADER, line...)
		case "DATA":
			conn.Write([]byte("354 start mail input\r\n"))
			for true {
				size, err = conn.Read(buf)
				data := buf[0:size]
				if size > 3 {
					x := data[size-3 : size]
					if string(x) == ".\r\n" {
						// data block end
						DATA = append(DATA, data[0:size-3]...)
						break
					} else {
						DATA = append(DATA, data...)
					}
				}
			}
			// DATA should be saved before sending OK
			conn.Write([]byte("250 OK\r\n"))
		case "QUIT":
			conn.Write([]byte("221 closing channel\r\n"))
			break

		}

	}
	fmt.Printf("HEADER IS: \n%s\n", string(HEADER))
	fmt.Printf("DATA IS: \n%s\n", string(DATA))
	msg := message{}
	msg.Header = string(HEADER)
	msg.Body = string(DATA)
	jsonbytes, err := json.Marshal(msg)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		var post bytes.Buffer
		post.Write(jsonbytes)
		_, err := httpskipssl.Post("http://localhost:9200/mails/inbox", "application/json", &post)
		if err != nil {
			fmt.Println(err.Error())
		}
	}
}

var httpskipssl = &http.Client{Transport: &http.Transport{
	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
}}
