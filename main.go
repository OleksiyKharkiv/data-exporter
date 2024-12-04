package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sort"
	"strings"
)

// ExecuteCommand executes a shell command and returns its output
func ExecuteCommand(command string) (string, error) {
	cmd := exec.Command("bash", "-c", command)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	return out.String(), err
}

// ConnectSSHWithPassphrase establishes an SSH connection using a private key.
func ConnectSSHWithPassphrase(remoteServer, privateKeyPath, user string) error {
	log.Println("Starting SSH connection...")

	// Step 1: Check if the private key file exists
	log.Printf("Checking if private key file exists at path: %s", privateKeyPath)
	if _, err := os.Stat(privateKeyPath); os.IsNotExist(err) {
		return fmt.Errorf("private key file not found: %s", privateKeyPath)
	}

	// Step 2: Read the private key file to ensure it is accessible
	log.Println("Reading private key file...")
	keyContent, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return fmt.Errorf("failed to read private key file: %w", err)
	}
	log.Printf("Private key file content read successfully. Size: %d bytes", len(keyContent))

	// Step 3: Format the SSH command for PowerShell
	command := fmt.Sprintf(`ssh -i "%s" %s@%s`, privateKeyPath, user, remoteServer)
	log.Printf("Formatted SSH command: %s", command)

	// Step 4: Execute the SSH command using PowerShell
	log.Println("Executing SSH command using PowerShell...")
	cmd := exec.Command("powershell", "-Command", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start SSH connection: %w", err)
	}

	log.Println("SSH connection established successfully.")
	return nil
}

// GetLatestDumpFolder retrieves the most recent "dump_" folder from remote server2
func GetLatestDumpFolder(remoteServer2, user2, password2 string) (string, error) {
	log.Println("Getting the latest dump folder...")

	// Use sshpass to connect via SSH and list directories
	command := fmt.Sprintf(`sshpass -p '%s' ssh -o StrictHostKeyChecking=no %s@%s "ls /data"`, password2, user2, remoteServer2)
	output, err := ExecuteCommand(command)
	if err != nil {
		return "", fmt.Errorf("failed to list dump folders: %w", err)
	}

	// Parse and filter the directory list
	lines := strings.Split(output, "\n")
	var dumpFolders []string
	for _, line := range lines {
		if strings.HasPrefix(line, "dump_") {
			dumpFolders = append(dumpFolders, line)
		}
	}

	if len(dumpFolders) == 0 {
		return "", fmt.Errorf("no dump folders found")
	}

	// Sort folders to find the latest one
	sort.Strings(dumpFolders)
	latestDumpFolder := dumpFolders[len(dumpFolders)-1]
	log.Printf("Found latest dump folder: %s", latestDumpFolder)
	return fmt.Sprintf("/data/%s", latestDumpFolder), nil
}

// FetchDumpSubfolders downloads specific subfolders from the latest dump folder
func FetchDumpSubfolders(remoteServer2, latestDumpFolder, localPath, user2, password2 string) ([]string, error) {
	log.Println("Fetching subfolders from the latest dump folder...")
	targetFolders := []string{"mats-user", "mats-payment", "mats-diagnostic", "mats-training-plan"}
	var successfulFolders []string //nolint:prealloc

	for _, folder := range targetFolders {
		remotePath := fmt.Sprintf("%s/%s", latestDumpFolder, folder)
		command := fmt.Sprintf(`sshpass -p '%s' sftp -r -P 30222 %s@%s:%s %s`, password2, user2, remoteServer2, remotePath, localPath)
		output, err := ExecuteCommand(command)
		if err != nil {
			log.Printf("Failed to fetch folder %s: %v\nOutput: %s", folder, err, output)
			continue
		}
		successfulFolders = append(successfulFolders, folder)
	}

	log.Printf("Successfully fetched subfolders: %v", successfulFolders)
	return successfulFolders, nil
}

func CleanOldDumps(remoteServer, user, privateKeyPath, remotePath string) error {
	// PowerShell command to connect and remove all folders using sudo
	psCommand := fmt.Sprintf(
		`ssh -i "%s" %s@%s "sudo find %s -mindepth 1 -maxdepth 1 -exec rm -r {} \;"`,
		privateKeyPath, user, remoteServer, remotePath,
	)

	// Run PowerShell command
	cmd := exec.Command("powershell", "-Command", psCommand)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to clean old dumps: %v\nOutput: %s", err, string(output))
	}

	log.Printf("Old dumps cleaned successfully. Output: %s", string(output))
	return nil
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile) // Logging with date and file
	log.Println("Starting data export...")

	remoteServer1 := "172.201.121.48"
	user1 := "MATS-VM-01_admin"
	privateKeyPath := "C:/MATS/Olex/Olex.pem"
	remotePath := "/var/lib/mongodb/download/dump/"
	remoteServer2 := "116.202.11.250"
	user2 := "jens"
	password2 := "J7Beuv0YI9qQCMY" //nolint:gosec

	// Step 1: Clean old dumps on remote server 1
	log.Println("Step 1: Cleaning old dumps on remote server 1...")
	err := CleanOldDumps(remoteServer1, user1, privateKeyPath, remotePath)
	if err != nil {
		log.Fatalf("Error cleaning old dumps: %v", err)
	}

	// Step 2: Get the latest dump folder from remote server 2
	log.Println("Step 2: Fetching the latest dump folder from remote server 2...")
	latestDumpFolder, fetchDumpErr := GetLatestDumpFolder(remoteServer2, user2, password2)
	if fetchDumpErr != nil {
		log.Printf("Critical error: Unable to find the latest dump folder on remote server 2: %v", fetchDumpErr)
		log.Println("Terminating the program. Ensure the remote server is reachable and the credentials are correct.")
		os.Exit(1)
	}

	log.Printf("Successfully identified the latest dump folder: %s", latestDumpFolder)

	// Step 3: Fetch specific subfolders from the latest dump folder
	log.Println("Step 3: Fetching subfolders from the latest dump folder...")
	successfulFolders, fetchSubfoldersErr := FetchDumpSubfolders(remoteServer2, latestDumpFolder, ".", user2, password2)
	if fetchSubfoldersErr != nil {
		log.Printf("Error fetching subfolders from the latest dump folder: %v", fetchSubfoldersErr)
		log.Println("Attempting to proceed with any subfolders that were successfully fetched.")
	}

	if len(successfulFolders) == 0 {
		log.Println("No subfolders were successfully fetched. This is a critical error.")
		log.Println("Terminating program as no data could be transferred.")
		os.Exit(1)
	}

	// Logging successful subfolders
	log.Println("Successfully fetched the following folders:")
	for _, folder := range successfulFolders {
		log.Println("- " + folder)
	}
}
