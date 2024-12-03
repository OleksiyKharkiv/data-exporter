package main

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
)

// Helper function to execute shell commands
func runCommand(cmd string, args []string, password string) (string, error) {
	command := exec.Command(cmd, args...)
	if password != "" {
		// Inject password into the command if necessary
		command.Stdin = bytes.NewBufferString(password + "\n")
	}
	output, err := command.CombinedOutput()
	return string(output), err
}

// Connect to the remote server via SSH
func sshConnect(user, server, pemFile, password string) error {
	fmt.Printf("Connecting to server %s@%s...\n", user, server)
	cmd := "ssh"
	args := []string{"-i", pemFile, fmt.Sprintf("%s@%s", user, server)}
	output, err := runCommand(cmd, args, password)
	if err != nil {
		return fmt.Errorf("failed to connect: %v, output: %s", err, output)
	}
	fmt.Println("Connection successful.")
	return nil
}

// Navigate and clean up the dump directory
func cleanupDumpDir(server, password string) error {
	fmt.Println("Navigating and cleaning dump directory...")
	cmd := "ssh"
	args := []string{server, "cd /var/lib/mongodb/download/dump && sudo rm -r *"}
	output, err := runCommand(cmd, args, password)
	if err != nil {
		return fmt.Errorf("failed to clean dump directory: %v, output: %s", err, output)
	}
	fmt.Println("Dump directory cleaned successfully.")
	return nil
}

// Fetch latest dump directories and copy them
func fetchDumps(remoteServer, remotePath, localPath, password string) ([]string, error) {
	fmt.Println("Fetching latest dumps...")
	// Mock logic to get the freshest dump directory (implement real sorting logic as needed)
	latestDumpDir := "dump_2024-12-01T21:30:03Z"
	cmd := "sftp"
	args := []string{"-r", "-P", "30222", fmt.Sprintf("%s:%s/%s", remoteServer, remotePath, latestDumpDir), localPath}
	output, err := runCommand(cmd, args, password)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch dumps: %v, output: %s", err, output)
	}
	fmt.Println("Dumps fetched successfully.")
	return []string{"mats-user", "mats-payment", "mats-diagnostic", "mats-training-plan"}, nil
}

// Main program
func main() {
	// SSH connection details
	sshUser := "MATS-VM-01_admin"
	sshServer := "172.201.121.48"
	sshPemFile := "C:\\MATS\\Olex\\Olex.pem"
	sshPassword := "your_password_here"

	// SFTP details
	sftpServer := "116.202.11.250"
	remotePath := "/data"
	localPath := "./"
	sftpPassword := "J7Beuv0YI9qQCMY"

	// Step 1: SSH Connection
	if err := sshConnect(sshUser, sshServer, sshPemFile, sshPassword); err != nil {
		log.Fatalf("Error in SSH connection: %v", err)
	}

	// Step 2: Cleanup dump directory
	if err := cleanupDumpDir(fmt.Sprintf("%s@%s", sshUser, sshServer), sshPassword); err != nil {
		log.Fatalf("Error in cleaning dump directory: %v", err)
	}

	// Step 3: Fetch and copy dump data
	dumpDirs, err := fetchDumps(sftpServer, remotePath, localPath, sftpPassword)
	if err != nil {
		log.Fatalf("Error in fetching dumps: %v", err)
	}

	// Step 4: Display result
	fmt.Println("Dump directories successfully copied:")
	for _, dir := range dumpDirs {
		fmt.Printf("- %s\n", dir)
	}
}
