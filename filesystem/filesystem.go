package filesystem

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/webp"
)

type ImageInfo struct {
	Name           string
	FileSize       string
	FileSizeFormat string
	Width          string
	Height         string
	ImageFormat    string
	ModifiedTime   string
}

func CreateImageDecodeConfig(path string) (image.Config, string, error) {
	file, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
		return image.Config{}, "", err
	}
	defer file.Close()

	config, format, err := image.DecodeConfig(file)
	if err != nil {
		fmt.Println(err)
		return image.Config{}, "", err
	}
	return config, format, nil

}

func getFileInfo(path string) (os.FileInfo, error) {
	f, err := os.Stat(path)

	if err != nil {
		fmt.Println(err)
		return f, err
	}

	return f, nil
}

func formatFileSize(bytes int64) (string, string, error) {

	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)

	sizeString := fmt.Sprintf("%d", bytes)
	sizeFormat := "Bytes"
	switch {
	case bytes >= GB:
		sizeString = fmt.Sprintf("%.2f", float64(bytes)/float64(GB))
		sizeFormat = "GB"
	case bytes >= MB:
		sizeString = fmt.Sprintf("%.2f", float64(bytes)/float64(MB))
		sizeFormat = "MB"
	case bytes >= KB:
		sizeString = fmt.Sprintf("%.2f", float64(bytes)/float64(KB))
		sizeFormat = "KB"
	}

	return sizeString, sizeFormat, nil
}

func GetImageInfo(path string) (ImageInfo, error) {

	decodeConfig, format, err := CreateImageDecodeConfig(path)

	if err != nil {
		fmt.Println(err)
		return ImageInfo{}, err
	}

	fileInfo, err := getFileInfo(path)

	if err != nil {
		fmt.Println(err)
		return ImageInfo{}, err
	}

	sizeString, sizeFormat, err := formatFileSize(fileInfo.Size())

	if err != nil {
		fmt.Println(err)
		return ImageInfo{}, err
	}

	return ImageInfo{
		Name:           path,
		FileSize:       sizeString,
		FileSizeFormat: sizeFormat,
		Width:          fmt.Sprintf("%d", decodeConfig.Width),
		Height:         fmt.Sprintf("%d", decodeConfig.Height),
		ImageFormat:    format,
		ModifiedTime:   fileInfo.ModTime().Format("2006-01-02 15:04:05"),
	}, nil

}
