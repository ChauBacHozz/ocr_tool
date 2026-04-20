package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"net/http"
	"os"

	// "path/filepath"
	"strings"

	// Thêm thư viện gọi Dialog native của hệ điều hành
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	fitz "github.com/gen2brain/go-fitz"
	sqdialog "github.com/sqweek/dialog"
)

func convertPDFToImages(pdfPath string, outputDir string) ([]image.Image, error) {
	// 1. Mở file PDF
	doc, err := fitz.New(pdfPath)
	if err != nil {
		return nil, fmt.Errorf("không thể mở file PDF: %v", err)
	}
	defer doc.Close()

	var imgs []image.Image

	// 2. Tạo thư mục output nếu chưa có
	os.MkdirAll(outputDir, os.ModePerm)

	// 3. Lặp qua từng trang (Index bắt đầu từ 0)
	for n := 0; n < doc.NumPage(); n++ {
		// Trích xuất trang thành ảnh (mặc định độ phân giải khá cao, phù hợp cho OCR)
		img, err := doc.Image(n)
		if err != nil {
			return nil, fmt.Errorf("lỗi trích xuất trang %d: %v", n, err)
		}
		imgs = append(imgs, img)

		// // Tạo tên file ảnh (VD: page_1.jpg, page_2.jpg)
		// outputPath := filepath.Join(outputDir, fmt.Sprintf("page_%d.jpg", n+1))

		// // Tạo file trên ổ cứng
		// f, err := os.Create(outputPath)
		// if err != nil {
		// 	return nil, err
		// }

		// // Lưu ảnh dưới định dạng JPEG (Quality = 90)
		// err = jpeg.Encode(f, img, &jpeg.Options{Quality: 90})
		// f.Close()

		// if err != nil {
		// 	return nil, err
		// }

		// // Thêm vào danh sách kết quả
		// imagePaths = append(imagePaths, outputPath)
		// fmt.Printf(" Đã xuất trang %d thành %s\n", n+1, outputPath)
	}

	return imgs, nil
}

func extractTextFromMemory(img image.Image) (string, error) {
	// 1. Tạo một bộ đệm trên RAM
	var buf bytes.Buffer

	// 2. Mã hóa bức ảnh vào bộ đệm đó dưới dạng JPEG
	err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90})
	if err != nil {
		return "", fmt.Errorf("lỗi mã hóa ảnh trên RAM: %v", err)
	}

	// 3. Chuyển byte trong bộ đệm thành chuỗi Base64
	base64Str := base64.StdEncoding.EncodeToString(buf.Bytes())
	dataURI := fmt.Sprintf("data:image/jpeg;base64,%s", base64Str)

	// 4. Đóng gói Payload JSON chuẩn OpenAI Vision (Giống hệt cũ)
	payload := map[string]interface{}{
		"model": "glm-ocr",
		"messages": []map[string]interface{}{
			{
				"role": "user",
				"content": []map[string]interface{}{
					{
						"type": "text",
						"text": "",
					},
					{
						"type":      "image_url",
						"image_url": map[string]string{"url": dataURI},
					},
				},
			},
		},
		"temperature": 0.1,
		"max_tokens":  2048,
	}

	jsonPayload, _ := json.Marshal(payload)

	// 5. Bắn Request tới Llama Server
	resp, err := http.Post("http://127.0.0.1:8080/v1/chat/completions", "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", fmt.Errorf("lỗi kết nối server: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("server trả lỗi %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	json.Unmarshal(body, &result)

	choices, ok := result["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return "", fmt.Errorf("server không trả về text")
	}

	firstChoice := choices[0].(map[string]interface{})
	message := firstChoice["message"].(map[string]interface{})

	return message["content"].(string), nil
}

func runOCR(pdfPath string) (string, error) {
	fmt.Println("Đang xử lý:", pdfPath)

	var displayBuffer strings.Builder
	displayBuffer.WriteString("Bắt đầu mở file PDF...\n")

	fyne.Do(func() {
		resultArea.SetText(displayBuffer.String())
	})

	doc, err := fitz.New(pdfPath)
	if err != nil {
		return "", fmt.Errorf("không thể mở file PDF: %v", err)
	}
	defer doc.Close()

	numPages := doc.NumPage()
	var fullTextBuilder strings.Builder

	for i := 0; i < numPages; i++ {
		statusMsg := fmt.Sprintf("--- Đang quét trang %d/%d ---\n", i+1, numPages)
		displayBuffer.WriteString(statusMsg)

		fyne.Do(func() {
			resultArea.SetText(displayBuffer.String())
		})

		img, err := doc.Image(i)
		if err != nil {
			msg := fmt.Sprintf("[Lỗi trích xuất trang %d]: %v\n", i+1, err)
			displayBuffer.WriteString(msg)
			fyne.Do(func() {
				resultArea.SetText(displayBuffer.String())
			})
			continue
		}

		extractedText, err := extractTextFromMemory(img)
		if err != nil {
			msg := fmt.Sprintf("[Lỗi OCR trang %d]: %v\n", i+1, err)
			displayBuffer.WriteString(msg)
		} else {
			fullTextBuilder.WriteString(extractedText)
			fullTextBuilder.WriteString("\n\n---\n\n")
			displayBuffer.WriteString(extractedText)
			displayBuffer.WriteString("\n\n---\n\n")
		}

		fyne.Do(func() {
			resultArea.SetText(displayBuffer.String())
		})

		img = nil
	}

	return fullTextBuilder.String(), nil
}

func openFileDialog(w fyne.Window, btn *widget.Button) {
	// 1. Những lệnh UI này chạy trên Main Thread (vì openFileDialog được gọi từ sự kiện nhấn nút)
	pdfPath, err := sqdialog.File().Filter("PDF files", "pdf", "PDF").Load()
	if err != nil {
		fmt.Println("Xảy ra lỗi khi chọn file hoặc người dùng hủy:", err)
		return
	}

	// Vô hiệu hóa nút NGAY LẬP TỨC trên main thread
	btn.Disable()
	oldText := btn.Text
	btn.SetText("Đang xử lý OCR... Vui lòng đợi")

	// 2. Chỉ chạy xử lý nặng (OCR, ghi file) trong Goroutine
	go func() {
		// Đảm bảo nút được kích hoạt lại khi kết thúc
		defer func() {
			fyne.Do(func() {
				btn.SetText(oldText)
				btn.Enable()
			})
		}()

		mdText, err := runOCR(pdfPath)
		if err != nil {
			dialog.ShowError(fmt.Errorf("Lỗi OCR: %v", err), w)
			return
		}

		// Tạo thư mục output nếu chưa có
		outputDir := "./output"
		if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
			dialog.ShowError(fmt.Errorf("Không thể tạo thư mục output: %v", err), w)
			return
		}

		// Ghi file
		outputPath := outputDir + "/output.md"
		err = os.WriteFile(outputPath, []byte(mdText), 0644)
		if err != nil {
			dialog.ShowError(fmt.Errorf("Lỗi khi lưu file: %v", err), w)
			return
		}

		fmt.Println("[THÔNG BÁO] Đã thực hiện OCR xong", pdfPath)

		// Hiển thị thông báo hệ thống
		fyne.CurrentApp().SendNotification(fyne.NewNotification("OCR Hoàn tất", "Kết quả đã được lưu vào: "+outputPath))

		// Hiển thị hộp thoại trong ứng dụng
		dialog.ShowInformation("Thành công", "Đã thực hiện OCR xong và lưu tại:\n"+outputPath, w)
	}()
}
