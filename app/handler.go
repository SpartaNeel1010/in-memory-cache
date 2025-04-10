package main 

import (
	"net"
	"fmt"
	"strconv"
	"strings"
	"time"
	"os"

)
func handleConnection(conn net.Conn) {

	defer conn.Close()
	logger.Info("Client connected", "address", conn.RemoteAddr().String())

	// reader := bufio.NewReader(conn)

	for {
		// Read the command from the client
		buff := make([]byte, 1024)
		length, err := conn.Read(buff)
		if err != nil {
			logger.Error("Error reading connection data", "error", err)
			break
		}
		msg := string(buff[:length])
		

		logger.Info("Received message", "address", conn.RemoteAddr().String(), "message", string(msg))

		// Parse the RESP message
		parsed, err := parseRESP(msg)




		if err != nil {
			logger.Error("Failed to parse RESP message", "error", err)
			conn.Write([]byte("-ERR Parsing failed\r\n"))
			continue
		}

		// Print parsed result (should be ["ECHO", "hey"])
		fmt.Println(parsed[0])
		// logger.Info("Parsed result", "message", parsed)

		// Respond to the ECHO command (this is for example purposes)
		command:=strings.ToLower(parsed[0])
		if len(parsed)==0 {
			conn.Write([]byte("-ERR unknown command\r\n"))
		}else if command == "ping"{
			conn.Write([]byte("$4\r\nPONG\r\n"))

		}else if command == "echo" {
			conn.Write([]byte("$" + strconv.Itoa(len(parsed[1])) + "\r\n" + parsed[1] + "\r\n"))
		}else if command=="set"{
			if len(parsed)<3{
				conn.Write([]byte("-ERR unknown command\r\n"))
				continue
			}
			key:=parsed[1]
			value:=parsed[2]
			if len(parsed)==5 && parsed[3]=="px" {
				expTm,err :=strconv.Atoi(parsed[4])
				currentTime := time.Now().UnixMilli()
				if err!=nil{
					logger.Error("Expiration time not an integer")
					conn.Write([]byte("-ERR unknown command\r\n"))
					continue 
				}
				expireTS:=currentTime+int64(expTm)
				expTime[key]=int(expireTS)
			} 
			
			DB[key]=value
			conn.Write([]byte("+OK\r\n"))

		}else if command=="get"{
			if len(parsed)<2{
				conn.Write([]byte("-ERR unknown command\r\n"))
				continue
			}
			key:=parsed[1]
			value, exists := DB[key]
			if !exists{
				conn.Write([]byte("$-1\r\n"))
				continue
			}

			currentTime := time.Now().UnixMilli()
			expTm,exist:=expTime[key]
			
		
			if exist && (currentTime > (int64(expTm))) {
				delete(expTime,key)
				delete(DB, key)
				
				conn.Write([]byte("$-1\r\n"))
				continue 
			}


			conn.Write([]byte("$" + strconv.Itoa(len(value)) + "\r\n" + value + "\r\n"))
			

		}else if command=="config"{
			if len(parsed)<3{
				conn.Write([]byte("-ERR unknown command\r\n"))
				continue;
			}
			value:=strings.ToLower(parsed[2])
			var ans string
			if value=="dir"{
				ans=*dir
			}else if value=="dbfilename"{
				ans=*dbFileName
			}else{
				conn.Write([]byte("-ERR unknown command\r\n"))
				continue;

			}
			
			conn.Write([]byte("*2\r\n$" + strconv.Itoa(len(value)) + "\r\n" + value + "\r\n" +"$"+ strconv.Itoa(len(ans))+ "\r\n"+ans+ "\r\n"))

		}else if command == "keys" {
			if len(parsed) < 2 || parsed[1] != "*" {
				conn.Write([]byte("-ERR unknown command\r\n"))
				continue
			}
		
			// Reload data from RDB file to ensure latest keys are fetched
		
		
			// Return all keys in the database
			keys,err := getAllKeys()
			
			if err != nil {
				fmt.Printf("Error opening file: %v\n", err)
				os.Exit(1)
			}

		
			// Respond with RESP array
			response := fmt.Sprintf("*%d\r\n", len(keys))
			for _, key := range keys {
				response += fmt.Sprintf("$%d\r\n%s\r\n", len(key), key)
			}
			conn.Write([]byte(response))
		} else if command == "save" {
			conn.Write([]byte("+OK\r\n"))
		}else if command == "info"{
			if len(parsed)<2{
				conn.Write([]byte("-ERR unknown command\r\n"))
				continue
			}
			caseType:=strings.ToLower(parsed[1])
			if caseType == "replication"{
				
				response := fmt.Sprintf("role:%s\nmaster_replid:%s\nmaster_repl_offset:%d\n", role, master_replid, master_repl_offset)
				resp:=fmt.Sprintf("$%d\r\n%s\r\n", len(response), response)

				conn.Write([]byte(resp))
			}else{
				conn.Write([]byte("-ERR unknown command\r\n"))
			}
			
		} else if command == "replconf"{
			if role != "master"{
				conn.Write([]byte("-ERR unknown command\r\n"))
				continue
			}
			conn.Write([]byte("+OK\r\n"))
		}else if command=="psync" {

			if role != "master" || len(parsed)<3{
				conn.Write([]byte("-ERR unknown command\r\n"))
				continue
			}
			reqReplID := strings.ToLower(parsed[1])
			slaveOffset := strings.ToLower(parsed[2])
			if reqReplID=="?" && slaveOffset=="-1"{
				conn.Write([]byte("+FULLRESYNC " + master_replid +" 0\r\n"))
				// The rdb file for the master is assumed to be empty for this test and here I have to write code for sending the rdb file in resp format to the slave (replica)
				// Slave will send replconf two times and then send pysnc (partial sync ) with master id and offset(-1 initially) 
				// However, this is the first time master knows about the replica thus it will send whole rdb file and will start full resynchronization 
				//  Also, I don't know If I have to recieve some kind of reply from the slave node 
				// I don't know how I can read the rdb file and decode it and store it in resp format 

				continue
			}
			conn.Write([]byte("-ERR unknown command\r\n"))

		}else {
			conn.Write([]byte("-ERR unknown command\r\n"))
		}
	}
	

}