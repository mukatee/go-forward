package forwarder

import (
	"strconv"
	"net"
	"log"
	"time"
)

//SendTestMsg can be used for testing to send messages to the forwarder
//done channel is used to signal the caller when sending the message is done, so they can check for receiving
//host is the target hostname/ip where to send message to, port is port for the same host
//msg is the string to send
func SendTestMsg(done chan bool, host string, port int, msg string) {
	var addr = host+":"+strconv.Itoa(port)
	conn, err := net.Dial("tcp", addr)

	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Close()

	//convert the msg to a byte array for writing to socket
	bytes := []byte(msg)
	conn.Write(bytes)
	log.Printf("TC Sent: %s", "hello")

	//create a byte array of size 1024 as buffer for reading the response.
	//this assumes the sender will not send any more bytes, which would buffer overflow.
	//but i assume this function is only used for small tests, so there you go
	buff := make([]byte, 1024)
	n, _ := conn.Read(buff)
	log.Printf("TC received "+strconv.Itoa(n))
	log.Printf("TC Receive: %#v", buff[:n])
	//if you check the docs for time.Sleep it does not tell you timeunit it uses. but this is how all the examples do it.
	//yes, can i have better docs please?
	time.Sleep(10 * time.Millisecond)
	done <- true
	log.Println("TC client send done")
}

//SendTestMsgClose is the same as the SendTestMsg but it closes the connection when it has sent the msg
func SendTestMsgClose(done chan bool, host string, port int, msg string) {
	var addr = host+":"+strconv.Itoa(port)
	conn, err := net.Dial("tcp", addr)

	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Close()

	bytes := []byte(msg)
	conn.Write(bytes)
	log.Printf("TC Sent: %s", "hello")

	conn.Close()
	log.Printf("TC connection closed")
	done <- true
	log.Println("TC client send done")
}
