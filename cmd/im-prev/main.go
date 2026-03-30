// package main

// import (
// 	"bufio"
// 	"fmt"
// 	"io"
// 	"os"
// 	"strings"

// 	"bytes"
// 	"os/exec"

// 	"golang.org/x/term"

// 	tea "github.com/charmbracelet/bubbletea"
// 	lipgloss "github.com/charmbracelet/lipgloss"
// 	imprev "github.com/miabobia/im-prev"
// 	"github.com/miabobia/im-prev/config"
// )

// func imageBorderStyle(width, height int) lipgloss.Style {
// 	return lipgloss.NewStyle().
// 		Border(lipgloss.RoundedBorder()).
// 		Width(width).
// 		Height(height).
// 		// Width(config.C.ImageBorderWidth).
// 		// Height(config.C.ImageBorderHeight).
// 		Align(lipgloss.Center, lipgloss.Center)
// }

// func selectorBorderStyle(width, height int, m model) lipgloss.Style {
// 	// selectorBorderHeight := config.C.SelectorBorderHeight
// 	// if len(m.fileShortNames) < config.C.SelectorBorderHeight {
// 	// 	selectorBorderHeight = len(m.fileShortNames)
// 	// }
// 	return lipgloss.NewStyle().
// 		Border(lipgloss.RoundedBorder()).
// 		// Width(config.C.SelectorBorderWidth).
// 		// Height(selectorBorderHeight).
// 		Width(width).
// 		Height(height).
// 		Align(lipgloss.Left, lipgloss.Center)
// }

// func boldStyle() lipgloss.Style {
// 	return lipgloss.NewStyle().Bold(true)
// }

// type model struct {
// 	imEntries      []imprev.ImageEntry
// 	fileShortNames []string
// 	cursor         int
// 	layout         Layout
// 	chafaConfig    imprev.ChafaConfig
// }

// type Layout struct {
// 	ImageWidth     int
// 	ImageHeight    int
// 	SelectorWidth  int
// 	SelectorHeight int
// 	ChafaWidth     int
// 	ChafaHeight    int
// }

// func computeLayout(termW, termH int, cfg *config.Config) Layout {
// 	return Layout{
// 		ImageWidth:     int(float32(termW) * cfg.ImageWidthRatio),
// 		ImageHeight:    int(float32(termH) * cfg.ImageHeightRatio),
// 		SelectorWidth:  int(float32(termW) * cfg.SelectorWidthRatio),
// 		SelectorHeight: int(float32(termH) * cfg.SelectorWidthRatio),
// 		ChafaWidth:     int(float32(termW) * cfg.ImageWidthRatio),
// 		ChafaHeight:    int(float32(termH) * cfg.ImageHeightRatio),
// 	}
// }

// func initialModel(imEntries []imprev.ImageEntry) model {
// 	// Cache short names once
// 	fsn := make([]string, len(imEntries))
// 	for i, entry := range imEntries {
// 		parts := strings.Split(entry.Path, "/")
// 		fsn[i] = parts[len(parts)-1]
// 	}

// 	return model{
// 		imEntries:      imEntries,
// 		fileShortNames: fsn,
// 		cursor:         0,
// 	}
// }

// func (m model) Init() tea.Cmd {
// 	return tea.ClearScreen
// }

// func (m *model) loadImagesIfNeeded() {
// 	// Calculate which images should be loaded
// 	halfWay := config.C.MaxFiles / 2
// 	startRange := m.cursor - halfWay - config.C.FileViewBufferSize
// 	// endRange := m.cursor + halfWay

// 	// Clamp to valid indices
// 	if startRange < 0 {
// 		startRange = 0
// 	}
// 	// if endRange >= len(m.imEntries) {
// 	// 	endRange = len(m.imEntries) - 1
// 	// }

// 	// Update (only loads images that aren't already loaded)
// 	updated, _ := imprev.UpdateChafaImages(m.imEntries, startRange)
// 	m.imEntries = updated
// }

// func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
// 	switch msg := msg.(type) {

// 	case tea.WindowSizeMsg:
// 		m.layout = computeLayout(msg.Width, msg.Height, config.C)
// 		m.imEntries = imprev.GetImageEntries()
// 	// 	m.termWidth = msg.Width
// 	// 	m.termHeight = msg.Height

// 	case tea.KeyMsg:
// 		switch msg.String() {
// 		case "ctrl+c", "q", "Q":
// 			return m, tea.Quit

// 		case "up", "k":
// 			if m.cursor > 0 {
// 				m.cursor--
// 			}
// 			m.loadImagesIfNeeded()

// 		case "down", "j":
// 			if m.cursor < len(m.imEntries)-1 { // Changed
// 				m.cursor++
// 			}
// 			m.loadImagesIfNeeded()
// 		case "c", "C":
// 			copyImage(m.imEntries[m.cursor].Path) // Changed
// 		}
// 	}
// 	return m, nil
// }

// func copyImage(path string) error {
// 	// open image
// 	file, err := os.Open(path)
// 	if err != nil {
// 		fmt.Println(err)
// 		return err
// 	}
// 	defer file.Close()

// 	// Get the file size
// 	stat, err := file.Stat()
// 	if err != nil {
// 		fmt.Println(err)
// 		return err
// 	}

// 	// Read the file into a byte slice
// 	bs := make([]byte, stat.Size())
// 	_, err = bufio.NewReader(file).Read(bs)
// 	if err != nil && err != io.EOF {
// 		fmt.Println(err)
// 		return err
// 	}

// 	// copy image
// 	cmd := exec.Command("wl-copy", "--type", "image/png")
// 	cmd.Stdin = bytes.NewReader(bs)
// 	cmd.Run()
// 	return nil
// }

// func truncateText(s string, max int) string {
// 	runes := []rune(s)
// 	if len(runes) <= max {
// 		return s
// 	}
// 	return string(runes[len(runes)-max:])
// }

// func constructFileList(m model) ([]string, int) {
// 	if len(m.imEntries) <= config.C.MaxFiles { // Changed
// 		return m.fileShortNames, m.cursor
// 	}

// 	halfWay := config.C.MaxFiles / 2

// 	if m.cursor < halfWay {
// 		return m.fileShortNames[:config.C.MaxFiles], m.cursor
// 	}

// 	if m.cursor >= len(m.imEntries)-halfWay { // Changed
// 		cursorPos := m.cursor - (len(m.fileShortNames) - config.C.MaxFiles)
// 		return m.fileShortNames[len(m.fileShortNames)-config.C.MaxFiles:], cursorPos
// 	}

// 	top := m.cursor - halfWay
// 	bottom := m.cursor + halfWay
// 	return m.fileShortNames[top:bottom], halfWay
// }

// func (m model) View() string {
// 	title := "\n Select an image to preview :)"
// 	fileList := ""

// 	truncFiles, relativeCursor := constructFileList(m)
// 	for index, file := range truncFiles {
// 		cursor := "  "
// 		fileName := truncateText(file, config.C.MaxFileNameLength)
// 		if relativeCursor == index {
// 			cursor = " >"
// 			fileName = boldStyle().Render(fileName)
// 		}
// 		fileList += fmt.Sprintf("%s %s\n", cursor, fileName)
// 	}

// 	fileList = selectorBorderStyle(m.layout.SelectorWidth, m.layout.SelectorHeight, m).Render(fileList)
// 	listWidth := lipgloss.Width(fileList)
// 	centeredTitle := lipgloss.NewStyle().
// 		Width(listWidth).
// 		Align(lipgloss.Center).
// 		Render(title)

// 	fileList = centeredTitle + "\n" + fileList

// 	rightSide := ""
// 	if m.cursor >= 0 && m.cursor < len(m.imEntries) { // Changed
// 		image := imageBorderStyle(m.layout.ImageWidth, m.layout.ImageHeight).Render(m.imEntries[m.cursor].ChafaImage) // Changed
// 		caption := "(C)opy (Q)uit"

// 		rightSide = lipgloss.JoinVertical(
// 			lipgloss.Left,
// 			image,
// 			caption,
// 		)
// 	}

// 	return lipgloss.JoinHorizontal(
// 		lipgloss.Top,
// 		fileList,
// 		"    ",
// 		rightSide,
// 	)
// }

// func main() {
// 	//	TODO:
// 	// 	- if no arg passed run in current directory
// 	// 	- loading bar during chafa load (load images dynamically?)
// 	//	- extra selector options? open with default image viewer, zoom, delete?
// 	//  - gif support???
// 	//  - symbol change?
// 	//  - have communication functions exposed from imprev.go so we can change structs on the fly

// 	if err := config.Load("config/config.toml"); err != nil {
// 		fmt.Println(err.Error())
// 		return
// 	}

// 	if len(os.Args) != 2 {
// 		fmt.Println("Error: there should be one argument included\neg: im-prev ~/Downloads")
// 		return
// 	}

// 	if _, err := exec.LookPath("wl-copy"); err != nil {
// 		fmt.Fprintf(os.Stderr, "wl-copy not found: %v\n", err)
// 		os.Exit(1)
// 	}

// 	path := os.Args[1]

// 	fmt.Println(path)

// 	width, height, err := term.GetSize(int(os.Stdout.Fd()))

// 	// this is probably very bad but here it is
// 	layout := computeLayout(width, height, config.C)
// 	chafaCfg := imprev.buildChafaConfig(layout.ChafaWidth, layout.ChafaHeight)

// 	imEntries, err := imprev.GetImageEntries(path, tmpLayout.ChafaWidth, tmpLayout.ChafaHeight)

// 	if err != nil {
// 		fmt.Println(err.Error())
// 		return
// 	}
// 	p := tea.NewProgram(initialModel(imEntries, layout, chafaCfg))
// 	if _, err := p.Run(); err != nil {
// 		fmt.Printf("Alas, there's been an error: %v", err)
// 		os.Exit(1)
// 	}
// }
