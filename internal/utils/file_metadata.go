package utils

import (
	"slices"
	"fmt"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
)

func GetFormattedFileSize(fileHeader *multipart.FileHeader) string {
	size := fileHeader.Size

	const (
		_          = iota
		KB float64 = 1 << (10 * iota)
		MB
		GB
	)

	switch {
	case float64(size) >= GB:
		return fmt.Sprintf("%.2f GB", float64(size)/GB)
	case float64(size) >= MB:
		return fmt.Sprintf("%.2f MB", float64(size)/MB)
	case float64(size) >= KB:
		return fmt.Sprintf("%.0f KB", float64(size)/KB)
	default:
		return fmt.Sprintf("%d B", size)
	}
}


func GetFileType(fileHeader *multipart.FileHeader) string {
	// 1. MIME TYPE Detection from file content
	file, err := fileHeader.Open()
	if err != nil {
		return "Unknown"
	}
	defer file.Close()

	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		return "Unknown"
	}

	contentType := http.DetectContentType(buffer)

	// 2. Extension detection
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))

	// Combine both methods
	if strings.HasPrefix(contentType, "image/") || slices.Contains([]string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".svg", ".webp"}, ext) {
		return "Images"
	}
	if strings.HasPrefix(contentType, "video/") || slices.Contains([]string{".mp4", ".mov", ".avi", ".mkv", ".flv", ".wmv"},ext) {
		return "Videos"
	}
	if strings.HasPrefix(contentType, "audio/") || slices.Contains([]string{".mp3", ".wav", ".aac", ".ogg", ".flac"},ext){
		return "Audio"
	}
	if strings.HasPrefix(contentType, "application/pdf") ||
		strings.HasPrefix(contentType, "application/msword") ||
		strings.HasPrefix(contentType, "application/vnd.openxmlformats-officedocument") ||
		slices.Contains([]string{".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx", ".txt", ".csv"},ext) {
		return "Documents"
	}
	if slices.Contains([]string{".zip", ".rar", ".7z", ".tar", ".gz"},ext) {
		return "Compressed"
	}
	if slices.Contains([]string{".html", ".css", ".js", ".ts", ".jsx", ".json", ".xml"}, ext) {
		return "Code"
	}
	if strings.HasPrefix(contentType, "text/") {
		return "Text Files"
	}

	return "Others"
}

 
