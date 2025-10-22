package imprev

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/miabobia/im-prev/config"
)

type ImageEntry struct {
	Path       string
	ChafaImage string
}

func getFiles(path string) ([]string, error) {
	files := []string{}

	entries, err := os.ReadDir(path)
	if err != nil {
		return files, err
	}

	extensions := []string{".png", ".webp", ".jpg", ".jpeg", ".bmp"}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		for _, ext := range extensions {
			if strings.HasSuffix(strings.ToLower(name), ext) {
				files = append(files, path+"/"+name)
				break
			}
		}
	}

	return files, nil
}

func loadChafaImages(fileArray []string, startRange int) ([]ImageEntry, error) {
	var wg sync.WaitGroup
	imEntries := make([]ImageEntry, len(fileArray))

	// construct chafa command
	chafaPath, err := exec.LookPath("chafa")

	if err != nil {
		fmt.Println(err.Error())
		return imEntries, err
	}

	sizeFlag := "--size"
	sizeArg := "50x20"
	symbolsFlag := "--symbols"
	symbolsArg := "braille"

	config.C.MaxFiles

	// structs to store data
	for i, fileName := range fileArray {
		imEntries[i].Path = fileName
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			cmd := exec.Command(
				chafaPath,
				fileName,
				symbolsFlag,
				symbolsArg,
				sizeFlag,
				sizeArg,
			)
			stdout, _ := cmd.Output()
			imEntries[index].ChafaImage = string(stdout)
		}(i)
	}
	wg.Wait()
	return imEntries, nil
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	return false, err
}

func GetImageEntries(path string) ([]ImageEntry, error) {
	_, err := exists(path)

	if err != nil {
		fmt.Println(err.Error())
		return []ImageEntry{}, err
	}

	fileArray, err := getFiles(path)

	if err != nil {
		fmt.Println(err.Error())
		return []ImageEntry{}, err
	}

	imEntries, err := loadChafaImages(fileArray)

	if err != nil {
		fmt.Println(err.Error())
		return imEntries, nil
	}

	return imEntries, nil
}
