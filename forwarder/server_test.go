package forwarder

import (
	"testing"
	"time"
	"log"
	"net"
	"strconv"
	"io"
	"bufio"
)

//Tests forwarding with only the destination configured, not the mirror
func TestOne(t *testing.T) {
	Config.srcPort = 9999
	Config.dstPort = 10000
	Config.dstHost = "localhost"
	Config.mirrorUpPort = 0
	Config.mirrorUpHost = ""

	go StartServer()
	received := make (chan string)
	go StartTestServer("FWD", Config.dstPort, received)

	time.Sleep(100 * time.Millisecond)

	done := make(chan bool)
	go SendTestMsgClose(done, "localhost", Config.srcPort, "hello_t")
	log.Printf("TS: waiting on TC done channel")
	<- done
	log.Printf("TS: TC done channel responded")

	actual := <- received
	log.Printf("TS: channel received string:"+actual)
	expected := "hello_t"
	if actual != expected {
		t.Fatalf("Expected %s but got %s", expected, actual)
	} else {
		log.Printf("Happy? Expected %s and got %s", expected, actual)
	}
}

//tests forwarding with both a destination and a mirror
func TestTwo(t *testing.T) {
	Config.srcPort = 9999
	Config.dstPort = 10000
	Config.dstHost = "localhost"
	Config.mirrorUpPort = 12001
	Config.mirrorUpHost = "localhost"

	go StartServer()
	received := make (chan string)
	go StartTestServer("FWD", Config.dstPort, received)
	mirrored:= make (chan string)
	go StartTestServer( "MIR", Config.mirrorUpPort, mirrored)

	time.Sleep(100 * time.Millisecond)

	done := make(chan bool)
	go SendTestMsgClose(done, "localhost", Config.srcPort, "hello_t")
	log.Printf("TS: waiting on TC done channel")
	<- done
	log.Printf("TS: TC done channel responded")

	actual := <- received
	log.Printf("TS: channel received string:"+actual)
	expected := "hello_t"
	if actual != expected {
		t.Fatalf("Destination expected on %s but got %s", expected, actual)
	} else {
		log.Printf("Happy? Destination expected %s and got %s", expected, actual)
	}

	actual = <- mirrored
	log.Printf("TS: channel received string:"+actual)
	expected = "hello_t"
	if actual != expected {
		t.Fatalf("Mirror expected on %s but got %s", expected, actual)
	} else {
		log.Printf("Happy? Mirror expected %s and got %s", expected, actual)
	}
}

func StartTestServer(id string, port int, received chan string) {
	addr := "localhost:"+strconv.Itoa(port)
	log.Printf("TS: trying to start on addr="+addr)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf(id+" TS: Listen failed: %v", err)
	}
	defer listener.Close()
	log.Printf(id+" TS: Listening for connection")

	conn, err := listener.Accept()
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf(id+" TS: Got connection %v -> %v", conn.RemoteAddr(), conn.LocalAddr())
	defer conn.Close()
	buf := make([]byte, 1024)
	r := bufio.NewReader(conn)
//	w := bufio.NewWriter(conn)

	for {
		n, err := r.Read(buf)
		if n > 0 {
			//even with err being non-nil (such as EOF, we should still process data that was received)
			log.Printf(id+" TS: received data, n=%v", n)
			s := string(buf[:n])
			log.Printf(" TS: "+s)
			received <- s
		}
		switch err {
		case io.EOF:
			//this means the connection has been closed
			log.Printf(id+" TS: EOF received, connection closed")
			return
		case nil:
			//lets not break the for loop but keep forwarding the stream..
			break
		default:
			log.Fatalf("TS: Receive data failed:%s", err)
			return
		}
	}
}
