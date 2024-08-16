package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
)

var (
	connections = make(map[net.Conn]string) // Map to hold connections and usernames
	mu          sync.Mutex
)

func handleConnection(conn net.Conn) {
	defer conn.Close()

	// Prompt for username
	fmt.Fprintln(conn, "Enter your username:")
	scanner := bufio.NewScanner(conn)
	if !scanner.Scan() {
		return
	}
	username := scanner.Text()
	
	mu.Lock()
	connections[conn] = username
	mu.Unlock()

	fmt.Printf("%s has joined the chat\n", username)

	// Broadcast welcome message
	broadcastMessage(fmt.Sprintf("%s has joined the chat", username), conn)

	for scanner.Scan() {
		message := scanner.Text()
		if message == "/quit" {
			break
		} else if strings.HasPrefix(message, "/msg") {
			handlePrivateMessage(message, username)
		} else if strings.HasPrefix(message, "/file") {
			handleFileTransfer(message, conn)
		} else {
			broadcastMessage(fmt.Sprintf("%s: %s", username, message), conn)
		}
	}

	mu.Lock()
	delete(connections, conn)
	mu.Unlock()

	fmt.Printf("%s has left the chat\n", username)
	broadcastMessage(fmt.Sprintf("%s has left the chat", username), nil)
}

func broadcastMessage(message string, excludeConn net.Conn) {
	mu.Lock()
	defer mu.Unlock()
	for conn := range connections {
		if conn != excludeConn {
			_, _ = fmt.Fprintln(conn, message)
		}
	}
}

func handlePrivateMessage(message, sender string) {
	parts := strings.SplitN(message, " ", 3)
	if len(parts) < 3 {
		return
	}
	recipient := parts[1]
	msg := parts[2]

	mu.Lock()
	defer mu.Unlock()
	for conn, username := range connections {
		if username == recipient {
			_, _ = fmt.Fprintf(conn, "%s (private): %s\n", sender, msg)
			break
		}
	}
}

func handleFileTransfer(message string, conn net.Conn) {
	parts := strings.SplitN(message, " ", 2)
	if len(parts) < 2 {
		return
	}
	filePath := parts[1]

	file, err := os.Open(filePath)
	if err != nil {
		fmt.Fprintln(conn, "Error opening file:", err)
		return
	}
	defer file.Close()

	_, _ = fmt.Fprintln(conn, "Sending file:", filePath)
	fileScanner := bufio.NewScanner(file)
	for fileScanner.Scan() {
		_, _ = fmt.Fprintln(conn, fileScanner.Text())
	}
	_, _ = fmt.Fprintln(conn, "File transfer completed.")
}

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer listener.Close()

	fmt.Println("Server started on :8080")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		go handleConnection(conn)
	}
}
