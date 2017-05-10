package forwarder

import (
	"net"
	"log"
	"strconv"
	"bufio"
	"io"
	"os"
)

const LOGFILE = "gofwd.log"

//stopper exists the main listener loop when set to true
var stopper bool = false
var ddf *os.File
var duf *os.File

//StartServer starts the forwarding server, which listens for connections and forwards them according to configuration
//configuration is expected to be in the global Config object/struct/whatever
func StartServer() {
	//create a log file and configure logging to both standard output and the file
	if Config.logFile != "" {
		f, err := os.OpenFile(Config.logFile, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("Failed to open log file for writing: %v", err)
		}
		if !Config.logToConsole {
			log.SetOutput(io.MultiWriter(os.Stdout, f))
		} else {
			log.SetOutput(io.MultiWriter(f))
		}
	} else {
		if Config.logToConsole {
			log.SetOutput(io.MultiWriter(os.Stdout))
		}
	}

	if Config.dataDownFile != "" {
		f, err := os.OpenFile(Config.dataDownFile, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
		ddf = f
		if err != nil {
			log.Fatalf("Failed to open downstream data-log file for writing: %v", err)
		}
	}
	if Config.dataUpFile != "" {
		f, err := os.OpenFile(Config.dataUpFile, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
		duf = f
		if err != nil {
			log.Fatalf("Failed to open upstream data-log file for writing: %v", err)
		}
	}

	listener, err := net.Listen("tcp", "localhost:"+strconv.Itoa(Config.srcPort))
	if err != nil {
		log.Fatalf("Listen failed: %v", err)
	}
	//close the listener when this function exits
	defer listener.Close()
	debuglog("Listening for connection")

	for {
		if stopper {
			//exit if someone set the flag to stop forwarding
			break
		}
		mainConn, err := listener.Accept()
		debuglog("Got connection %v -> %v", mainConn.RemoteAddr(), mainConn.LocalAddr())
		if err != nil {
			//TODO; should probably not exit on error from one client..
			log.Fatalln(err)
		}
		//start a new thread for this connection and wait for the next one
		go forward(mainConn)
	}
}

//writes to log if logging to console or file is enabled
//honestly, I just copied the v... interface{} from the log package definition so there you go
func debuglog(msg string, v... interface{}) {
	if Config.logToConsole || Config.logFile != "" {
		log.Printf(msg, v...)
	}
}

//writes traffic data as string to log file if file is enabled
func filelog(msg string, v... interface{}) {
	if Config.logFile != "" {
		log.Printf(msg, v...)
	}
}

//forward() handles forwarding of a given source connection to configured destion/mirror
func forward(srcConn net.Conn) {
	//have to defer here as defer waits for surrounding function to return.
	//deferring in main() for loop would only execute when main() exits (?)
	defer srcConn.Close()

	//set up main destination, the one whose returned data is also written back to source connection
	addr := Config.dstHost + ":" + strconv.Itoa(Config.dstPort)
	dstConn, err := net.Dial("tcp", addr)

	if err != nil {
		//try not to fail on a single error when forwarding a single connection. maybe destination is down and will be up, or maybe there is temporary network outage etc?
		log.Printf("Connection to destination failed. Skipping connection. Error: %v", err)
		return
	}
	debuglog("Dialed %v -> %v", dstConn.LocalAddr(), dstConn.RemoteAddr())

	//set up the mirror connection if upstream mirroring is defined
	var mirrorUpConn net.Conn
	if Config.mirrorUpPort > 0 {
		addr := Config.mirrorUpHost + ":" + strconv.Itoa(Config.mirrorUpPort)
		mirrorUpConn, err = net.Dial("tcp", addr)
		if err != nil {
			debuglog("Connection to upstream mirror failed. Skipping mirror for this connection. Error: %v", err)
			mirrorUpConn = nil
		} else {
			debuglog("Connected to upstream mirror:"+addr)
		}
	}

	//set up the mirror connection if downstream mirroring is defined
	var mirrorDownConn net.Conn
	if Config.mirrorDownPort > 0 {
		addr := Config.mirrorDownHost + ":" + strconv.Itoa(Config.mirrorDownPort)
		mirrorDownConn, err = net.Dial("tcp", addr)
		if err != nil {
			debuglog("Connection to downstream mirror failed. Skipping mirror for this connection. Error: %v", err)
			mirrorDownConn = nil
		} else {
			debuglog("Connected to downstream mirror:"+addr)
		}
	}
	//create channels to wait for until the forwarding of upstream and downstream data is done.
	//these are needed to enable channel waits or the defer on the source connection close() executes immediately and breaks all stream forwards
	fwd1Done := make(chan bool)
	fwd2Done := make(chan bool)

	//forward the source data to destination and the mirror, and destination data to the source
	//only source -> destination traffic is mirrored. not destination -> source. just add the other part if you need
	go streamFwd(srcConn, dstConn, mirrorUpConn, "src->dst", fwd1Done, true)
	go streamFwd(dstConn, srcConn, mirrorDownConn, "dst->src", fwd2Done, false)
	//wait until the stream forwarders exit to exit this function so the srcConn.close() is not prematurely executed
	<- fwd1Done
	<- fwd2Done
}

//streamFwd forwards a given source connection to the given destination and mirror connections
//the id parameter is used to give more meaningful prints, and the done channel to report back when the forwarding ends
func streamFwd(srcConn net.Conn, dstConn net.Conn, mirrorConn net.Conn, id string, done chan bool, upstream bool) {
	defer srcConn.Close()
	defer dstConn.Close()
	r := bufio.NewReader(srcConn)
	w := bufio.NewWriter(dstConn)

	var mw *bufio.Writer
	if mirrorConn != nil {
		mw = bufio.NewWriter(mirrorConn)
		debuglog(id+": initializing with mirror")
	} else {
		debuglog(id+": initializing without mirror")
	}

	//buffer for reading data from source and forwarding it to the destination/mirror
	//notice that a separate call to this streamFwd() function is made for src->dst and dst->src so just need one buffer and one read/write pair here
	buf := make([]byte, Config.bufferSize)
LOOPER:
	for {
		n, err := r.Read(buf)
		if n > 0 {
			debuglog(id+": forwarding data, n=%v", n)
			w.Write(buf[:n])
			w.Flush()
			debuglog(id+": Write done")
			if mw != nil {
				mw.Write(buf[:n])
				mw.Flush()
				debuglog(id+": Writing to mirror done")
			}
		} else {
			debuglog(id+": no data received? n=%v", n)
		}
		if upstream {
			if duf != nil {
				//https://gobyexample.com/writing-files
				n2, err := duf.Write(buf[:n])
				duf.Sync()
				debuglog("wrote upstream data to file n=%v, err=%v", n2, err)
			}
		} else {
			if ddf != nil {
				n2, err := ddf.Write(buf[:n])
				ddf.Sync()
				debuglog("wrote downstream data to file n=%v, err=%v", n2, err)
			}
		}
		//%x would print the data in hex. could be made an option or whatever
//		debuglog("data=%x", data)

		switch err {
		case io.EOF:
			//this means the connection has been closed
			debuglog("EOF received, connection closed")
			break LOOPER
		case nil:
			//its a successful read, so lets not break the for loop but keep forwarding the stream..
			break
		default:
			//lets not crash the program on single socket error. better to wait for more connections to forward
			//log.Fatalf("Receive data failed:%s", err)
			debuglog("Breaking stream fwd due to error:%s", err)
			break LOOPER
		}

	}
	debuglog("exiting stream fwd")
	//notify the forward() function that this streamFwd() call has finished
	done <- true
}

//stops the listener waiting for new connections but does not interrupt existing ones
//did not test so issues quite likely :)
func stop() {
	stopper = true
}

