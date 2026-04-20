package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"
)

var (
	serverCmd *exec.Cmd
	mu        sync.Mutex
)

func getResourcePath() string {
	// 1. Ưu tiên 1: Thư mục làm việc hiện tại (tốt cho 'go run .' và chạy binary từ terminal)
	cwd, _ := os.Getwd()
	if _, err := os.Stat(filepath.Join(cwd, "bin", "llama-server")); err == nil {
		return cwd
	}

	// 2. Ưu tiên 2: Thư mục chứa file thực thi (tốt cho file binary được copy đi chỗ khác)
	exePath, _ := os.Executable()
	exeDir := filepath.Dir(exePath)
	if _, err := os.Stat(filepath.Join(exeDir, "bin", "llama-server")); err == nil {
		return exeDir
	}

	// 3. Ưu tiên 3: Nếu là macOS .app, bin thường nằm cùng cấp với folder .app hoặc trong Resources
	// Thử tìm ở thư mục cha của thư mục chứa exe (thường là Contents/MacOS -> lùi lại 2 cấp)
	parentDir := filepath.Dir(filepath.Dir(exeDir)) // Ra khỏi MacOS và Contents
	bundleDir := filepath.Dir(parentDir)            // Ra khỏi .app
	if _, err := os.Stat(filepath.Join(bundleDir, "bin", "llama-server")); err == nil {
		return bundleDir
	}

	return cwd // Mặc định trả về cwd nếu không tìm thấy
}

func startServer() error {
	mu.Lock()
	defer mu.Unlock()

	baseDir := getResourcePath()

	var binary string

	switch runtime.GOOS {
	case "windows":
		binary = filepath.Join(baseDir, "bin", "llama-b8851-bin-win-cpu-x64", "llama-server.exe")
	case "darwin":
		binary = filepath.Join(baseDir, "bin", "llama-b8838-mac", "llama-server")
	default:
		return fmt.Errorf("hệ điều hành không được hỗ trợ: %s", runtime.GOOS)
	}

	modelPath := filepath.Join(baseDir, "models", "LightOnOCR-2-1B-Q4_K_M.gguf")
	mmprojPath := filepath.Join(baseDir, "models", "mmproj-LightOnOCR-2-1B-Q8_0.gguf")

	if _, err := os.Stat(binary); err != nil {
		return fmt.Errorf("không tìm thấy llama-server tại: %s. Vui lòng đảm bảo thư mục 'bin' nằm cùng cấp với ứng dụng.", binary)
	}

	serverCmd = exec.Command(binary,
		"-m", modelPath,
		"--mmproj", mmprojPath,
		"--port", "8080",
		"-c", "6000",
		"-t", "1",
		"-fa", "on",
		"--parallel", "1",
	)

	serverCmd.Stdout = os.Stdout
	serverCmd.Stderr = os.Stderr

	err := serverCmd.Start()
	if err != nil {
		return err
	}

	fmt.Printf("🚀 Llama Server (100%% Local) đã bật - PID: %d\n", serverCmd.Process.Pid)
	return nil
}

func stopServer() {
	mu.Lock()
	defer mu.Unlock()

	if serverCmd != nil && serverCmd.Process != nil {
		fmt.Println("Chuẩn bị tắt LLama server")
		err := serverCmd.Process.Signal(syscall.SIGTERM)
		if err != nil {
			// Đã sửa thành Printf để in lỗi chuẩn xác
			fmt.Printf("Lỗi khi tắt server: %v\n", err)
			serverCmd.Process.Kill()
		}
		serverCmd.Wait()
		fmt.Println("Đã tắt Llama server")
	}
}
