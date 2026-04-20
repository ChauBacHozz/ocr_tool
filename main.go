package main

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	// sqdialog "github.com/sqweek/dialog"
)

var resultArea *widget.Entry

func main() {
	if err := startServer(); err != nil {
		fmt.Println("Không thể khởi động server:", err)
		return
	}
	defer stopServer()
	// --- Giao diện Fyne ---
	myApp := app.New()
	myWindow := myApp.NewWindow("Traphaco macOS OCR")

	status := widget.NewLabel("Hệ thống OCR đang chạy ngầm...")
	
	resultArea = widget.NewMultiLineEntry()
	resultArea.SetPlaceHolder("Kết quả OCR sẽ hiển thị ở đây...")

	var btn *widget.Button
	btn = widget.NewButton("Upload pdf files", func() {
		openFileDialog(myWindow, btn)
	})
	btn.Importance = widget.SuccessImportance

	content := container.NewBorder(
		status,
		btn,
		nil, nil,
		container.NewScroll(resultArea),
	)

	myWindow.SetContent(content)

	myWindow.Resize(fyne.NewSize(600, 400))
	myWindow.ShowAndRun()
}
