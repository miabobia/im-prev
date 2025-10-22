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

var chafaCfg ChafaConfig

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

type ChafaConfig struct {
	ChafaPath string
	SizeArg   string
	SymbolArg string
}

func buildChafaConfig() (ChafaConfig, error) {
	chafaPath, err := exec.LookPath("chafa")
	if err != nil {
		return ChafaConfig{}, fmt.Errorf("chafa not found: %w", err)
	}

	return ChafaConfig{
		ChafaPath: chafaPath,
		SizeArg:   fmt.Sprintf("%dx%d", config.C.ChafaMaxWidth, config.C.ChafaMaxHeight),
		SymbolArg: config.C.ChafaDefaultSymbols,
	}, nil
}

func getChafaImage(cfg ChafaConfig, fileName string) (string, error) {
	cmd := exec.Command(
		cfg.ChafaPath,
		fileName,
		"--size", cfg.SizeArg,
		"--symbols", cfg.SymbolArg,
	)
	stdout, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("chafa failed for %s: %w", fileName, err)
	}
	return string(stdout), nil
}

func UpdateChafaImages(imEntries []ImageEntry, startRange int) ([]ImageEntry, error) {
	// cfg, err := buildChafaConfig()
	// if err != nil {
	// 	return nil, err
	// }

	var wg sync.WaitGroup

	endRange := startRange + config.C.MaxFiles + config.C.FileViewBufferSize
	if endRange >= len(imEntries) {
		endRange = len(imEntries) - 1
	}
	for i := startRange; i < endRange; i++ {
		if imEntries[i].ChafaImage != "" {
			continue
		}
		wg.Add(1)

		go func(index int, file string) {
			defer wg.Done()

			chafaImage, err := getChafaImage(chafaCfg, file)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			imEntries[index].ChafaImage = chafaImage
		}(i, imEntries[i].Path)
	}
	// for i, imEntry := range imEntries {
	// 	fileName := imEntry.Path
	// 	if i < startRange || i > endRange {
	// 		// outside of range so we should clear it and not doing processing otherwise
	// 		imEntries[i].ChafaImage = ""
	// 		continue
	// 	}

	// 	if imEntries[i].ChafaImage != "" {
	// 		continue
	// 	}
	// 	wg.Add(1)

	// 	go func(index int, file string) {
	// 		defer wg.Done()

	// 		chafaImage, err := getChafaImage(cfg, file)
	// 		if err != nil {
	// 			fmt.Println(err.Error())
	// 			return
	// 		}
	// 		imEntries[index].ChafaImage = chafaImage
	// 	}(i, fileName)
	// }

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

func initImageEntries(fileArray []string) []ImageEntry {
	imEntries := make([]ImageEntry, len(fileArray))
	for i, filePath := range fileArray {
		imEntries[i].Path = filePath
	}
	return imEntries
}

func GetImageEntries(path string) ([]ImageEntry, error) {
	_, err := exists(path)

	if err != nil {
		fmt.Println(err.Error())
		return []ImageEntry{}, err
	}

	chafaCfg, _ = buildChafaConfig()

	fileArray, err := getFiles(path)

	if err != nil {
		fmt.Println(err.Error())
		return []ImageEntry{}, err
	}

	imEntries := initImageEntries(fileArray)

	imEntries, err = UpdateChafaImages(imEntries, 0)

	if err != nil {
		fmt.Println(err.Error())
		return imEntries, nil
	}

	return imEntries, nil
}
