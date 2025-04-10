package main

import (
	"flag"
	"log/slog"
	"net"
	"os"
	"fmt"


)
var port = flag.String("port", "6379", "Port to listen on")
var dir = flag.String("dir", "/tmp/redis-data", "Directory to store RDB file")
var dbFileName = flag.String("dbfilename", "dump.rdb", "RDB file name")
var replicaOf = flag.String("replicaof", "nil", "Is this server a replica? If yes then what is host Id and port number")



var _ = net.Listen

var _ = os.Exit
var logger *slog.Logger = slog.New(slog.NewTextHandler(os.Stderr, nil))
var fullPath string;

// info replication
var role string="master"
var master_repl_offset int =0
var master_replid string ="8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb"



var(
 DB map[string]string = make(map[string]string)
 expTime map[string]int = make(map[string]int)
 
)
func createPath() {
	// Check if file exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		fmt.Printf("File %s does not exist. Creating directory and file...\n", fullPath)

		
		fmt.Printf("Attempting to create directory: %s\n", *dir)

		// Create the directory if it doesn't exist
		err := os.MkdirAll(*dir, os.ModePerm)
		if err != nil {
			fmt.Printf("Failed to create directory %s: %v\n", *dir, err)
			return
		}
		fmt.Printf("Directory %s created successfully.\n", *dir)

		// Log file creation attempt
		fmt.Printf("Attempting to create file: %s\n", fullPath)

		// Create the file
		file, err := os.Create(fullPath)
		if err != nil {
			fmt.Printf("Failed to create file %s: %v\n", fullPath, err)
			return
		}
		defer file.Close()

		fmt.Printf("File %s created successfully.\n", fullPath)
	} else if err != nil {
		// If os.Stat() fails with a different error, log it
		fmt.Printf("Error checking file %s: %v\n", fullPath, err)
	}
}

func main() {
	flag.Parse()
	
	fullPath= *dir + "/" + *dbFileName

	if(*replicaOf != "nil"){
		role="slave"
		connectToMaster()
	}

	createPath()
	loadDB()

	
	host := "localhost"

	l, err := net.Listen("tcp","0.0.0.0:"+*port)
	if err != nil {
		logger.Error("Failed to bind to port ", "Port",*port)
		os.Exit(1)
	}
	logger.Info("creating server", "port", *port, "host", host)
	defer l.Close()

	



	for {
		conn, err := l.Accept()
		logger.Info("Accepted connection ", "remote address", conn.RemoteAddr().String())
		if err != nil {
			logger.Error("Error accepting connection", "error", err.Error())
			os.Exit(1)
		}
		go handleConnection(conn)
	}
	
}