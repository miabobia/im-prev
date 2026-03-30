package filesystem

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
)

type ImageInfo struct {
	Name           string
	FileSize       string
	FileSizeFormat string
	Width          string
	Height         string
	ImageFormat    string
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

func getFileSize(path string) (string, string, error) {

	f, err := os.Stat(path)

	if err != nil {
		fmt.Println(err)
		return "", "", err
	}

	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)

	size := f.Size()
	sizeString := fmt.Sprintf("%d", size)
	sizeFormat := "Bytes"
	switch {
	case size >= GB:
		sizeString = fmt.Sprintf("%.2f", float64(size)/float64(GB))
		sizeFormat = "GB"
	case size >= MB:
		sizeString = fmt.Sprintf("%.2f", float64(size)/float64(MB))
		sizeFormat = "MB"
	case size >= KB:
		sizeString = fmt.Sprintf("%.2f", float64(size)/float64(KB))
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

	sizeString, sizeFormat, err := getFileSize(path)

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
	}, nil

}
