
package main

import (
	"fmt"
)

func sendMessageToReplicas(message string) {
	
	// Print length of replicaConnections	
	fmt.Printf("Number of replica connections: %d\n", len(replicaConnections))
	
	for _, conn := range replicaConnections {
		// Send the message to each replica connection
		// Check if connection is open or not
		if conn == nil {
			fmt.Println("Connection is nil, skipping...")
			continue
		}
		fmt.Printf("Sending message to replica: %s\n", message)
		_, err := conn.Write([]byte(message))
		if err != nil {
			fmt.Printf("Failed to send message to replica: %v\n", err)
		}

	}
	
}


func temp(){
	// Iterate over all the connections in replicaConnections and send them "*3\r\n$8\r\nreplconf\r\n$6\r\ngetack\r\n$1\r\n*\r\n" message
	
	for _, conn := range replicaConnections {	
		if conn == nil {
			fmt.Println("Connection is nil, skipping...")
			continue
		}
		message := "*3\r\n$8\r\nreplconf\r\n$6\r\ngetack\r\n$1\r\n*\r\n"
		fmt.Printf("Sending message to replica: %s\n", message)
		_, err := conn.Write([]byte(message))
		if err != nil {
			fmt.Printf("Failed to send message to replica: %v\n", err)
		}
	}
	// Print the number of replica connections
}

