package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

// main function: handles SSH connection and clean-up of old dumps
func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile) // Logging with date and file
	log.Println("Starting data export...")

	// Define variables
	remoteServer1 := "172.201.121.48"
	user1 := "MATS-VM-01_admin"
	privateKeyPath := "C:/MATS/Olex/Olex.pem"
	remotePath := "/var/lib/mongodb/download/dump/"

	// Step 1: Check if private key file exists
	log.Printf("Checking if private key file exists at path: %s", privateKeyPath)
	if _, err := os.Stat(privateKeyPath); os.IsNotExist(err) {
		log.Fatalf("Private key file not found: %s", privateKeyPath)
	}

	// Step 2: Read the private key file to ensure it is accessible
	log.Println("Reading private key file...")
	keyContent, err := os.ReadFile(privateKeyPath)
	if err != nil {
		log.Fatalf("Failed to read private key file: %v", err)
	}
	log.Printf("Private key file content read successfully. Size: %d bytes", len(keyContent))

	// Step 3: Format the SSH command for PowerShell
	command := fmt.Sprintf(`ssh -i "%s" %s@%s "sudo find %s -mindepth 1 -maxdepth 1 -exec rm -r {} \;"`, privateKeyPath, user1, remoteServer1, remotePath)
	log.Printf("Formatted SSH command: %s", command)

	// Step 4: Execute the SSH command using PowerShell
	log.Println("Executing SSH command using PowerShell...")
	cmd := exec.Command("powershell", "-Command", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		log.Fatalf("Failed to execute SSH command: %v", err)
	}

	log.Println("SSH connection established and old dumps cleaned successfully.")
}
