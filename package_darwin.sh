#!/bin/bash

# Package the application for macOS
echo "Packaging application..."
fyne package -os darwin -icon icon.png -name "Traphaco OCR" --app-id com.traphaco.ocr65

# Check if the package command succeeded
if [ $? -eq 0 ]; then
    echo "Package created successfully. Copying bin and models folders..."
    
    # Copy toàn bộ nội dung bên trong bin và models vào MacOS
    cp -R bin "Traphaco OCR.app/Contents/MacOS/"
    cp -R models "Traphaco OCR.app/Contents/MacOS/"
    
    echo "Done! Folders copied to Traphaco OCR.app/Contents/MacOS"
else
    echo "Error: fyne package command failed."
    exit 1
fi
