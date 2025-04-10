package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

)

func connectToMaster(){
	// getting address of master node from which replication is to be done 
	arr := strings.Split(*replicaOf," ")
	host:=arr[0]
	port2:=arr[1]
	address:=host + ":" + port2




	// connecting to master node
	conn, err := net.Dial("tcp", address)
	if err != nil {
		fmt.Println("Connection error:", err)
		return
	}
	defer conn.Close()
	fmt.Println("Connected to", address)



	// Send the "PING" message
	_, err = conn.Write([]byte("*1\r\n$4\r\nPING\r\n"))
	if err != nil {
		fmt.Println("Error sending data:", err)
		return
	}
	fmt.Println("PING sent")




	// Waiting for response of Ping message (+PONG)
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return
	}
	fmt.Println("Received:", string(buf[:n]))





	// Sending REPLCONF message  two times
	confCommand := "*3\r\n$8\r\nREPLCONF\r\n$14\r\nlistening-port\r\n" + "$" + strconv.Itoa(len(*port)) + "\r\n" + *port + "\r\n"
	_, err = conn.Write([]byte(confCommand))
	if err != nil {
		fmt.Println("Error sending REPLCONF listening-port:", err)
		return
	}

	conn.SetReadDeadline(time.Now().Add(5 * time.Second))

	buf = make([]byte, 1024)
	n, err = conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return
	}

	fmt.Println("Received:", string(buf[:n]))







	_, err = conn.Write([]byte("*3\r\n$8\r\nREPLCONF\r\n$4\r\ncapa\r\n$6\r\npsync2\r\n"))
	if err != nil {
		fmt.Println("Error sending REPLCONF capa:", err)
		return
	}

	fmt.Println("REPLCONF commands sent")

	// Waiting for OK message 
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))

	// Read the response from the server
	buf = make([]byte, 1024)
	n, err = conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return
	}
	fmt.Println("Received:", string(buf[:n]))






	_, err = conn.Write([]byte("*3\r\n$5\r\nPSYNC\r\n$1\r\n?\r\n$2\r\n-1\r\n"))
	if err != nil {
		fmt.Println("Error sending PSYNC:", err)
		return
	}
	fmt.Println("PSYNC done")



}