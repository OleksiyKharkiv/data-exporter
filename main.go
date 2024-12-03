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

func ConnectSSHWithPassphrase(remoteServer, privateKeyPath, user string) (*exec.Cmd, error) {
	log.Println("Starting SSH connection...")

	// Проверяем существование файла .pem
	if _, err := os.Stat(privateKeyPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("private key file not found: %s", privateKeyPath)
	}

	// Формируем SSH-команду с флагом -tt
	command := fmt.Sprintf(`ssh -tt -i "%s" %s@%s`, privateKeyPath, user, remoteServer)

	// Создаём команду для выполнения
	cmd := exec.Command("bash", "-c", command)

	// Перенаправляем ввод-вывод
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	// Логируем команду перед запуском
	log.Printf("Running SSH command: %s", command)

	// Запускаем и ожидаем завершения
	err := cmd.Run()
	if err != nil {
		log.Printf("Error starting SSH command: %v", err)
		return nil, fmt.Errorf("failed to start SSH connection: %v", err)
	}

	log.Println("SSH connection established.")
	return cmd, nil
}

// GetLatestDumpFolder retrieves the most recent "dump_" folder in the /data directory
func GetLatestDumpFolder(remoteServer, user, password string) (string, error) {
	log.Println("Getting the latest dump folder...")

	// Используем sshpass для подключения через SSH и получения списка папок
	command := fmt.Sprintf(`sshpass -p '%s' ssh -o StrictHostKeyChecking=no %s@%s "ls /data"`, password, user, remoteServer)
	output, err := ExecuteCommand(command)
	if err != nil {
		return "", fmt.Errorf("failed to list dump folders: %v", err)
	}

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

	// Сортируем папки по имени и выбираем самую последнюю
	sort.Strings(dumpFolders)
	latestDumpFolder := dumpFolders[len(dumpFolders)-1]
	log.Printf("Found latest dump folder: %s", latestDumpFolder)
	return fmt.Sprintf("/data/%s", latestDumpFolder), nil
}

// FetchDumpSubfolders downloads specific subfolders from the latest dump folder
func FetchDumpSubfolders(remoteServer, latestDumpFolder, localPath, user, password string) ([]string, error) {
	log.Println("Fetching subfolders from the latest dump folder...")
	targetFolders := []string{"mats-user", "mats-payment", "mats-diagnostic", "mats-training-plan"}
	var successfulFolders []string

	for _, folder := range targetFolders {
		remotePath := fmt.Sprintf("%s/%s", latestDumpFolder, folder)
		command := fmt.Sprintf(`sshpass -p '%s' sftp -r -P 30222 %s@%s:%s %s`, password, user, remoteServer, remotePath, localPath)
		_, err := ExecuteCommand(command)
		if err != nil {
			log.Printf("Failed to fetch folder %s: %v\n", folder, err)
			continue
		}
		successfulFolders = append(successfulFolders, folder)
	}

	log.Printf("Successfully fetched subfolders: %v", successfulFolders)
	return successfulFolders, nil
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile) // Enable timestamps and file/line number logging
	log.Println("Starting data export...")

	remoteServer := "116.202.11.250"
	user := "jens"
	password := "J7Beuv0YI9qQCMY"
	privateKeyPath := "C:/MATS/Olex/Olex.pem"

	log.Println("Step 1: Connecting to the remote server via SSH...")
	_, err := ConnectSSHWithPassphrase(remoteServer, privateKeyPath, user)
	if err != nil {
		log.Printf("Error connecting to remote server: %v\n", err)
		return
	}

	log.Println("Step 2: Fetching the latest dump folder...")
	latestDumpFolder, fetchDumpErr := GetLatestDumpFolder(remoteServer, user, password)
	if fetchDumpErr != nil {
		log.Fatalf("Error finding latest dump folder: %v", fetchDumpErr) // Завершаем выполнение с логированием
	}
	log.Printf("Latest dump folder: %s", latestDumpFolder)

	log.Println("Step 3: Fetching subfolders from the latest dump folder...")
	successfulFolders, fetchSubfoldersErr := FetchDumpSubfolders(remoteServer, latestDumpFolder, ".", user, password)
	if fetchSubfoldersErr != nil {
		log.Fatalf("Error fetching subfolders: %v", fetchSubfoldersErr) // Завершаем выполнение с логированием
	}

	log.Println("Successfully fetched the following folders:")
	for _, folder := range successfulFolders {
		log.Println("- " + folder)
	}
}
