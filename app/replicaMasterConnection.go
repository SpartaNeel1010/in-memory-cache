package main
import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)
var(
	offset int = 0
	
)

func connectAndHandleMaster() {
	// getting address of master node from which replication is to be done
	arr := strings.Split(*replicaOf, " ")
	host := arr[0]
	port2 := arr[1]
	address := host + ":" + port2
	isConnectedToMaster = true
	
	fmt.Println("Connecting to master at", address)
	// connecting to master node
	masterConn, err := net.Dial("tcp", address)
	if err != nil {
		fmt.Println("Connection error:", err)
		return
	}
	fmt.Println("Connected to", address)
	
	// Set a read deadline for the handshake phase
	// masterConn.SetReadDeadline(time.Now().Add(5 * time.Second))
	
	reader := bufio.NewReader(masterConn)
	state := 0
	for {
		switch state {
		case 0:
			// Send PING command
			_, err = masterConn.Write([]byte("*1\r\n$4\r\nPING\r\n"))
			if err != nil {
				fmt.Println("Error sending PING:", err)
				return
			}
			fmt.Println("PING sent")
			
			// Read PING response
			buf := make([]byte, 1024)
			n, err := reader.Read(buf)
			if err != nil {
				fmt.Println("Error reading PING response:", err)
				return
			}
			fmt.Println("Received:", string(buf[:n]))
			state++
		case 1:
			// Send first REPLCONF command (listening-port)
			confCommand := "*3\r\n$8\r\nREPLCONF\r\n$14\r\nlistening-port\r\n$" + strconv.Itoa(len(*port)) + "\r\n" + *port + "\r\n"
			_, err = masterConn.Write([]byte(confCommand))
			if err != nil {
				fmt.Println("Error sending REPLCONF listening-port:", err)
				return
			}
			
			buf := make([]byte, 1024)
			n, err := reader.Read(buf)
			if err != nil {
				fmt.Println("Error reading REPLCONF listening-port response:", err)
				return
			}
			fmt.Println("Received:", string(buf[:n]))
			state++
		case 2:
			// Send second REPLCONF command (capa)
			_, err = masterConn.Write([]byte("*3\r\n$8\r\nREPLCONF\r\n$4\r\ncapa\r\n$6\r\npsync2\r\n"))
			if err != nil {
				fmt.Println("Error sending REPLCONF capa:", err)
				return
			}
			
			
			buf := make([]byte, 1024)
			n, err := reader.Read(buf)
			if err != nil {
				fmt.Println("Error reading REPLCONF capa response:", err)
				return
			}
			fmt.Println("Received:", string(buf[:n]))
			state++
		case 3:
			// Send PSYNC command
			_, err = masterConn.Write([]byte("*3\r\n$5\r\nPSYNC\r\n$1\r\n?\r\n$2\r\n-1\r\n"))
			if err != nil {
				fmt.Println("Error sending PSYNC:", err)
				return
			}
			
			// Using the previously declared buffered reader
			
			// Read the FULLRESYNC response line
			fullResyncLine, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println("Error reading FULLRESYNC response:", err)
				return
			}
			fmt.Println("Received:", fullResyncLine)
			
			// Read the RDB header line (should start with '$')
			rdbHeader, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println("Error reading RDB header:", err)
				return
			}
			fmt.Println("Received RDB header:", rdbHeader)
			
			if len(rdbHeader) == 0 || rdbHeader[0] != '$' {
				fmt.Println("Unexpected RDB header:", rdbHeader)
				return
			}
			
			// Parse the RDB file size from the header
			rdbSize, err := strconv.Atoi(strings.TrimSpace(rdbHeader[1:]))
			if err != nil {
				fmt.Println("Error parsing RDB size:", err)
				return
			}
			
			// Read the entire RDB file data
			rdbData := make([]byte, rdbSize)
			totalRead := 0
			for totalRead < rdbSize {
				n, err := reader.Read(rdbData[totalRead:])
				if err != nil {
					fmt.Println("Error reading RDB data:", err)
					return
				}
				totalRead += n
			}
			
			fmt.Println("RDB file received")
			
			state++
		default:
			// Handle continuous messages from the master
			buf := make([]byte, 1024)
			n, err := reader.Read(buf)
			if err != nil {
				fmt.Println("Error reading message from master:", err)
				return
			}
			msg := string(buf[:n])

			logger.Info("Received message from master", "message", msg)
			
			parsed, err := parseRESP(msg)
			// print parsed here as an array 
			logger.Info("Parsed result", "message", parsed)
			if err != nil {
				logger.Error("Failed to parse RESP message", "error", err)
				masterConn.Write([]byte("-ERR Parsing failed\r\n"))
				continue
			}
			command := strings.ToLower(parsed[0])
			if command == "ping" {
				fmt.Println("PING received from master: Master is Alive")
				
				if( len(parsed) > 1) {
					parsed = parsed[1:]
					command= strings.ToLower(parsed[0])
				}
				offset+=14
			}
			for command == "set" {

				if len(parsed) < 3 {
					// logger.Error("ERR unknown command")
					continue
				}

				offset += 3 + len(parsed[1]) + len(parsed[2]) + 22 +
					(len(strconv.Itoa(len(parsed[1]))) - 1) +
					(len(strconv.Itoa(len(parsed[2]))) - 1)

				key := parsed[1]
				value := parsed[2]
				if len(parsed) == 5 && parsed[3] == "px" {
					expTm, err := strconv.Atoi(parsed[4])
					currentTime := time.Now().UnixMilli()
					if err != nil {
						logger.Error("Expiration time not an integer")
						continue
					}
					expireTS := currentTime + int64(expTm)
					expTime[key] = int(expireTS)
				}
				fmt.Println("Replica: Setting key:", key, "to value:", value)
				DB[key] = value

				
				if( len(parsed) > 3) {
					parsed = parsed[3:]
					command= strings.ToLower(parsed[0])
				}else {
					break
				}
			}
			if  command == "replconf"{
				// receieving command like REPLCONF GETACK *
				
				// write on masterConn with REPLCONG ACK offset
				fmt.Println("Replica: Received REPLCONF command")
				if len(parsed)==3 && strings.ToLower(parsed[1])=="getack" {
					fmt.Println("Replica: Received REPLCONF GETACK *")
			
					offsetStr := strconv.Itoa(offset)
					ackCommand := "*3\r\n$8\r\nREPLCONF\r\n$3\r\nACK\r\n$" + strconv.Itoa(len(offsetStr)) + "\r\n" + offsetStr + "\r\n"
					fmt.Println(ackCommand)
					_, err = masterConn.Write([]byte(ackCommand))
					if err != nil { 
						fmt.Println("Error sending REPLCONF ACK:", err)
						return
					}
					fmt.Println("REPLCONF ACK sent")
				}
				offset += 37
				
			}
		}
	}
}