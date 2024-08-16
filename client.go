package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	conn   net.Conn
	reader *bufio.Reader
	input  string
}

func (m model) Init() tea.Cmd {
	return tea.Batch(tea.EnterAltScreen)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case tea.KeyMsg:
		if strings.TrimSpace(m.input) != "" {
			if _, err := fmt.Fprintln(m.conn, m.input); err != nil {
				fmt.Println("Error sending message:", err)
				return m, tea.Quit
			}
			m.input = ""
		}
	case tea.Msg:
		if response, err := m.reader.ReadString('\n'); err == nil {
			fmt.Print(response)
		}
	}
	return m, nil
}

func (m model) View() string {
	return "Type your message and press Enter to send. Type /quit to exit.\n" + m.input
}

func main() {
	// Connect to server
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	defer conn.Close()

	fmt.Println("Connected to server")

	// Read server welcome message
	scanner := bufio.NewScanner(conn)
	if scanner.Scan() {
		fmt.Println(scanner.Text())
	}

	// Create a new Bubble Tea model
	m := model{
		conn:   conn,
		reader: bufio.NewReader(conn),
	}

	go func() {
		for {
			// Read messages from server
			response, err := m.reader.ReadString('\n')
			if err != nil {
				fmt.Println("Error reading from server:", err)
				return
			}
			fmt.Print(response)
		}
	}()

	// Start Bubble Tea program
	p := tea.NewProgram(m)
	if err := p.Start(); err != nil {
		fmt.Println("Error starting Bubble Tea:", err)
	}
}
