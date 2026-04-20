#!/bin/bash

# 1. Cài đặt mingw-w64 nếu chưa có (chỉ cần chạy 1 lần)
# brew install mingw-w64

echo "Packaging application for Windows..."
export PATH=$PATH:~/go/bin 

# Cấu hình biên dịch chéo cho Windows 64-bit
CGO_ENABLED=1 \
CC=x86_64-w64-mingw32-gcc \
GOOS=windows \
GOARCH=amd64 \
fyne package -os windows -icon icon.png -name "Traphaco OCR"

# Kiểm tra nếu thành công
if [ $? -eq 0 ]; then
    echo "Package created successfully. Preparing distribution folder..."
    
    mkdir -p dist_windows
    
    mv "Traphaco OCR.exe" dist_windows/ 2>/dev/null || mv "Traphaco OCR.zip" dist_windows/
    
    cp -R bin dist_windows/
    cp -R models dist_windows/
    
    # Zip thư mục dist_windows
    echo "Zipping dist_windows..."
    zip -r "Traphaco_OCR_Windows.zip" dist_windows/
    
    if [ $? -eq 0 ]; then
        echo "Done! Zip file created: Traphaco_OCR_Windows.zip"
    else
        echo "Error: Failed to create zip file."
        exit 1
    fi

    echo "Note: Make sure bin/ and models/ are NOT empty before copying."
else
    echo "Error: fyne package for windows failed."
    echo "Tip: Ensure you have mingw-w64 installed (brew install mingw-w64)."
    exit 1
fi