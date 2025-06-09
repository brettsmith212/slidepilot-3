package main

import (
	"fmt"
	"net"
	"os/exec"
	"time"
)

// StartLibreOfficeHeadless starts LibreOffice in headless mode with UNO socket
func StartLibreOfficeHeadless() error {
	// Check if LibreOffice is already running on port 8100
	if isPortOpen("127.0.0.1:8100") {
		fmt.Println("LibreOffice headless already running on port 8100")
		return nil
	}

	fmt.Println("Starting LibreOffice headless service...")
	
	cmd := exec.Command("soffice", 
		"--headless", 
		"--invisible", 
		"--nodefault", 
		"--nolockcheck", 
		"--nologo", 
		"--norestore",
		"--accept=socket,host=127.0.0.1,port=8100;urp;StarOffice.ServiceManager")
	
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start LibreOffice: %v", err)
	}

	// Wait for the service to be ready
	for i := 0; i < 10; i++ {
		if isPortOpen("127.0.0.1:8100") {
			fmt.Println("LibreOffice headless service ready")
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}

	return fmt.Errorf("LibreOffice headless service failed to start")
}

// StopLibreOfficeHeadless stops the LibreOffice headless service
func StopLibreOfficeHeadless() error {
	fmt.Println("Stopping LibreOffice headless service...")
	cmd := exec.Command("pkill", "-f", "soffice.*headless")
	return cmd.Run()
}

// isPortOpen checks if a port is open
func isPortOpen(address string) bool {
	conn, err := net.DialTimeout("tcp", address, 500*time.Millisecond)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}
