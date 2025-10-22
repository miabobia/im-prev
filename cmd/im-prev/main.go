package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	lipgloss "github.com/charmbracelet/lipgloss"
	imprev "github.com/miabobia/im-prev"
	"github.com/miabobia/im-prev/config"
	clipboard "golang.design/x/clipboard"
)

// var imageBorderStyle = lipgloss.NewStyle().
// 	Border(lipgloss.RoundedBorder()).
// 	Width(config.C.ImageBorderWidth).
// 	Height(config.C.ImageBorderHeight).
// 	Align(lipgloss.Center, lipgloss.Center)

// var selectorBorderStyle = lipgloss.NewStyle().
// 	Border(lipgloss.RoundedBorder()).
// 	Width(config.C.SelectorBorderWidth).
// 	Height(config.C.SelectorBorderHeight).
// 	Align(lipgloss.Left, lipgloss.Center)

// var boldStyle = lipgloss.NewStyle().Bold(true)

func imageBorderStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Width(config.C.ImageBorderWidth).
		Height(config.C.ImageBorderHeight).
		Align(lipgloss.Center, lipgloss.Center)
}

func selectorBorderStyle(m model) lipgloss.Style {
	selectorBorderHeight := config.C.SelectorBorderHeight
	if len(m.fileShortNames) < config.C.SelectorBorderHeight {
		selectorBorderHeight = len(m.fileShortNames)
	}
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Width(config.C.SelectorBorderWidth).
		Height(selectorBorderHeight).
		Align(lipgloss.Left, lipgloss.Center)
}

func boldStyle() lipgloss.Style {
	return lipgloss.NewStyle().Bold(true)
}

type model struct {
	fileChoices    []string
	fileShortNames []string
	chafaImages    []string
	cursor         int
}

func initialModel(imEntries []imprev.ImageEntry) model {

	fChoices := make([]string, len(imEntries))
	cImages := make([]string, len(imEntries))

	for i, imEntry := range imEntries {
		fChoices[i] = imEntry.Path
		cImages[i] = imEntry.ChafaImage
	}

	fsn := make([]string, len(fChoices))
	for index, f := range fChoices {
		parts := strings.Split(f, "/")
		fsn[index] = parts[len(parts)-1]
	}

	return model{
		fileChoices:    fChoices,
		fileShortNames: fsn,
		chafaImages:    cImages,
		cursor:         0,
	}
}

func (m model) Init() tea.Cmd {
	return tea.ClearScreen
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// Is it a key press?
	case tea.KeyMsg:

		// Cool, what was the actual key pressed?
		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c", "q", "Q":
			return m, tea.Quit

		// The "up" and "k" keys move the cursor up
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		// The "down" and "j" keys move the cursor down
		case "down", "j":
			if m.cursor < len(m.fileChoices)-1 {
				m.cursor++
			}

		case "c", "C":
			copyImage(m.fileChoices[m.cursor])
		}
	}
	return m, nil
}

func copyImage(path string) error {
	// open image
	file, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer file.Close()

	// Get the file size
	stat, err := file.Stat()
	if err != nil {
		fmt.Println(err)
		return err
	}

	// Read the file into a byte slice
	bs := make([]byte, stat.Size())
	_, err = bufio.NewReader(file).Read(bs)
	if err != nil && err != io.EOF {
		fmt.Println(err)
		return err
	}

	// copy image
	clipboard.Write(clipboard.FmtImage, bs)
	return nil
}

func truncateText(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[len(runes)-max:])
}

func constructFileList(m model) ([]string, int) {
	if len(m.fileChoices) <= config.C.MaxFiles {
		return m.fileShortNames, m.cursor
	}

	halfWay := config.C.MaxFiles / 2

	// cursor is near top of list still doesn't need to be shifted
	if m.cursor < halfWay {
		return m.fileShortNames[:config.C.MaxFiles], m.cursor
	}

	// cursor is near bottom of list
	if m.cursor >= len(m.fileChoices)-halfWay {
		cursorPos := m.cursor - (len(m.fileShortNames) - config.C.MaxFiles)
		return m.fileShortNames[len(m.fileShortNames)-config.C.MaxFiles:], cursorPos
	}

	// files need to be shifted
	top := m.cursor - halfWay
	bottom := m.cursor + halfWay
	return m.fileShortNames[top:bottom], halfWay
}

func (m model) View() string {

	title := "\n Select an image to preview :)"
	fileList := ""

	truncFiles, relativeCursor := constructFileList(m)
	for index, file := range truncFiles {
		cursor := "  "
		fileName := truncateText(file, config.C.MaxFileNameLength)
		if relativeCursor == index {
			cursor = " >"
			fileName = boldStyle().Render(fileName)
		}
		fileList += fmt.Sprintf("%s %s\n", cursor, fileName)
	}

	fileList = selectorBorderStyle(m).Render(fileList)

	// center title over selector
	listWidth := lipgloss.Width(fileList)
	centeredTitle := lipgloss.NewStyle().
		Width(listWidth).
		Align(lipgloss.Center).
		Render(title)

	fileList = centeredTitle + "\n" + fileList
	// Build the image preview (right side)
	rightSide := ""
	if m.cursor >= 0 && m.cursor < len(m.chafaImages) {
		image := imageBorderStyle().Render(m.chafaImages[m.cursor])
		caption := "(C)opy (Q)uit"

		// Stack image and text vertically
		rightSide = lipgloss.JoinVertical(
			lipgloss.Left, // alignment
			image,
			caption,
		)
	}

	// Join them horizontally
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		fileList,
		"    ",
		rightSide,
	)
}

func main() {
	//	TODO:
	// 	- if no arg passed run in current directory
	// 	- loading bar during chafa load (load images dynamically?)
	//	- extra selector options? open with default image viewer, zoom, delete?
	//  - gif support???
	//  - symbol change?
	//  - have communication functions exposed from imprev.go so we can change structs on the fly

	if err := config.Load("config/config.toml"); err != nil {
		fmt.Println(err.Error())
		return
	}

	if len(os.Args) != 2 {
		fmt.Println("Error: there should be one argument included\neg: im-prev ~/Downloads")
		return
	}

	if err := clipboard.Init(); err != nil {
		fmt.Println(err.Error())
		return
	}

	path := os.Args[1]

	fmt.Println(path)

	imEntries, err := imprev.GetImageEntries(path)

	if err != nil {
		fmt.Println(err.Error())
		return
	}
	p := tea.NewProgram(initialModel(imEntries))
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
