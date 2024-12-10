package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

// main function: handles SSH connection and clean-up of old dumps
//
//nolint:funlen
func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile) // Logging with date and file
	log.Println("Starting data export...")

	// Define variables
	remoteServer1 := "172.201.121.48"
	user1 := "MATS-VM-01_admin"
	privateKeyPath := "/root/.ssh/Olex.pem"
	remotePath := "/var/lib/mongodb/download/dump/"

	remoteServer2 := "116.202.11.250"
	user2 := "jens"
	password2 := "J7Beuv0YI9qQCMY" //nolint:gosec

	// Step 1: Check if a private key file exists
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

	// Step 3: Format the SSH command for PowerShell (now Unix)
	command := fmt.Sprintf("ssh -i \"%s\" %s@%s \"sudo find %s -mindepth 1 -maxdepth 1 -exec rm -r {} \\;\"", privateKeyPath, user1, remoteServer1, remotePath)
	log.Printf("Formatted SSH command: %s", command)

	// Step 4: Execute the SSH command to clean old dumps
	log.Println("Executing SSH command to clean old dumps...")
	cmd := exec.Command("bash", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		log.Fatalf("Failed to execute SSH command: %v", err)
	}
	log.Println("Old dumps cleaned successfully.")

	// Step 5: Create a temporary script for SFTP
	log.Println("Creating temporary script for SFTP...")
	sftpScript := fmt.Sprintf(`
		latest_dump=$(sshpass -p "%s" sftp -oPort=30222 -i "%s" %s@%s <<EOF
		ls /data/
		bye
		EOF
		| grep 'dump_' | sort -r | head -n 1 | awk '{print $1}')
		echo "Latest dump: $latest_dump"
	`, password2, privateKeyPath, user2, remoteServer2)

	// Write SFTP script to a temporary file
	tempScriptPath := "./sftp-script.sh"
	if err := os.WriteFile(tempScriptPath, []byte(sftpScript), 0644); err != nil { //nolint:gosec
		log.Fatalf("Failed to create SFTP script: %v", err)
	}
	log.Printf("Temporary script created at: %s", tempScriptPath)

	// Step 6: Execute the SFTP script
	log.Println("Executing SFTP script...")
	cmd = exec.Command("bash", tempScriptPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		log.Fatalf("Failed to execute SFTP script: %v", err)
	}
	log.Println("SFTP script executed successfully.")
}
