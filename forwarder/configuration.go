package forwarder

import (
	"os"
	"fmt"
	"flag"
)

//Configuration for the forwarder. Since it is capitalized, should be accessible outside package.
type Configuration struct {
	srcPort int //source where incoming connections to forward are listened to
	dstPort int //port on destination host where to forward incoming data/connections to. data coming back from destination is forwarded back using the source connection.
	dstHost string //destination host ip/name
	mirrorUpPort int //port on mirror host where to also forward incoming data/connection upstream traffic (from source to destination)
	mirrorUpHost string //upstream mirror host ip/name
	mirrorDownPort int //port on mirror host where to also forward incoming data/connection downstream traffic (from destination to source)
	mirrorDownHost string //downstream mirror host ip/name
	dataDownFile string  //if defined, write downstream data passed through to this file
	dataUpFile string  //if defined, write upstream data passed through to this file
	logFile string  //if defined, write log to this file
	logToConsole bool  //if we should log to console
	bufferSize int //size to use for buffering read/write data
}

//this is how go defines variables, so the actual configurations are stored here
var Config Configuration

//ParseConfig reads the command line arguments and sets the global Configuration object from those. Also checks the arguments make basic sense.
func ParseConfig() {
	flagSet := flag.NewFlagSet("goforward", flag.ExitOnError)
	flagSet.SetOutput(os.Stdout)

	srcPortPtr := flagSet.Int("sp", 0,"Source port for incoming connections. Required.")
	//here I was trying to create a shorter form of source port config option, while having the previous one be longer. the default package does not really support that, so dropped it.
	//	flagSet.IntVar(srcPortPtr, "sp", -1, "Source port for incoming connections.")

	dstPortPtr := flagSet.Int("dp", 0, "Destination port to forward incoming connections. Required.")
	dstHostPtr := flagSet.String("dh", "", "Destination host to forward incoming connections. Required.")

	mirrorUpPortPtr := flagSet.Int("mup", 0, "Mirror port to forward incoming connection upstream data. Optional. Required if upstream mirror host is defined.")
	mirrorUpHostPtr := flagSet.String("muh", "", "Mirror host to forward incoming connection upstream traffic. Optional.")

	mirrorDownPortPtr := flagSet.Int("mdp", 0, "Mirror port to forward incoming connection downstream data. Optional. Required if downstream mirror host is defined.")
	mirrorDownHostPtr := flagSet.String("mdh", "", "Mirror host to forward incoming connection downstream traffic. Optional.")

	dataUpFilePtr := flagSet.String("duf", "","If defined, will write upstream data to this file.")
	dataDownFilePtr := flagSet.String("ddf", "","If defined, will write downstream data to this file.")

	logFilePtr := flagSet.String("logf", "","If defined, will write debug log info to this file.")
	logToConsolePtr := flagSet.Bool("logc", false, "If defined, write debug log info to console.")
	bufferSizePtr := flagSet.Int("bufs", 1024, "Size of read/write buffering.")

	//the first argument is the name of the executable, so if the command was launched without parameters, print the help
	if len(os.Args) == 1 {
		fmt.Println("Usage: "+os.Args[0]+" [options]")
		fmt.Println(" Options:")
		flagSet.PrintDefaults()
		os.Exit(0)
	}

	flagSet.Parse(os.Args[1:])

	Config.srcPort = *srcPortPtr
	Config.dstPort = *dstPortPtr
	Config.dstHost = *dstHostPtr
	Config.mirrorUpPort = *mirrorUpPortPtr
	Config.mirrorUpHost = *mirrorUpHostPtr
	Config.mirrorDownPort = *mirrorDownPortPtr
	Config.mirrorDownHost = *mirrorDownHostPtr
	Config.dataDownFile = *dataDownFilePtr
	Config.dataUpFile = *dataUpFilePtr
	Config.logFile = *logFilePtr
	Config.logToConsole = *logToConsolePtr
	Config.bufferSize = *bufferSizePtr

	var errors = ""
	if Config.srcPort < 1 || Config.srcPort > 65535 {
		errors += "You need to specify source port in range 1-65535.\n"
	}
	if len(Config.dstHost) == 0 {
		errors += "You need to specify destination host.\n"
	}
	if Config.dstPort < 1 || Config.dstPort > 65535 {
		errors += "You need to specify destination port in range 1-65535.\n"
	}
	if Config.bufferSize < 1 {
		errors += "Buffer size needs to be >= 1.\n"
	}
	if len(Config.mirrorUpHost) > 0 {
		if Config.mirrorUpPort < 1 || Config.mirrorUpPort > 65535 {
			errors += "When upstream mirror host is defined, its port must be defined in range 1-65535.\n"
		}
	} else {
		if Config.mirrorUpPort != 0 {
			errors += "Mirror-up port defined but no mirror-up host. Mirror host is required if mirror is enabled.\n"
		}
	}
	if len(Config.mirrorDownHost) > 0 {
		if Config.mirrorDownPort < 1 || Config.mirrorDownPort > 65535 {
			errors += "When downstream mirror host is defined, its port must be defined in range 1-65535.\n"
		}
	} else {
		if Config.mirrorDownPort != 0 {
			errors += "Mirror-down port defined but no mirror-down host. Mirror host is required if mirror is enabled.\n"
		}
	}

	if len(errors) > 0 {
		fmt.Print(errors)
		fmt.Println()
		fmt.Print("Usage: goforward [options]")
		flagSet.PrintDefaults()
		os.Exit(1)
	}
}


