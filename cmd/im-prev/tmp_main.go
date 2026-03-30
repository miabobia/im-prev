// TODO
// search bar
// info section
// directory jumping ?
// recursive directory search
// help
// version
// filename truncating
// current directory

package main

import (
	"fmt"
	"os"
	"os/exec"

	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"

	// local packages
	imprev "github.com/miabobia/im-prev"
	"github.com/miabobia/im-prev/clipboard"
	"github.com/miabobia/im-prev/config"
	"github.com/miabobia/im-prev/filesystem"
	"github.com/miabobia/im-prev/styles"
)

type model struct {
	entries          []Entry
	cursor           int
	layout           Layout
	chafaPath        string
	clipboardManager clipboard.ClipboardManager
	searchMode       bool
	infoMode         bool
}

type Layout struct {
	ImageWidth     int
	ImageHeight    int
	SelectorWidth  int
	SelectorHeight int
	ChafaWidth     int
	ChafaHeight    int
	MaxFiles       int
}

type LoadState int

const (
	Unloaded LoadState = iota
	Loading
	Loaded
)

type Entry struct {
	data      string
	name      string
	loadState LoadState
}

// Tea messages

type imageLoadedMsg struct {
	index int
	image string
}

type imageUnloadedMsg struct {
	index int
	image string
}

type imageLoadErrorMsg struct {
	index int
	err   error
}

func loadImageCmd(chafaPath string, layout Layout, entry Entry, index int) tea.Cmd {
	return func() tea.Msg {
		chafaImage, _ := imprev.GetChafaImage(
			config.C.ChafaDefaultSymbols,
			fmt.Sprintf("%dx%d", layout.ChafaWidth, layout.ChafaHeight),
			chafaPath,
			entry.name,
		)
		return imageLoadedMsg{index: index, image: chafaImage}
	}
}

func unloadImageCmd(entry Entry, index int) tea.Cmd {
	return func() tea.Msg {
		return imageUnloadedMsg{index: index}
	}
}

// generic helper func
func Map[T, U any](s []T, f func(T) U) []U {
	result := make([]U, len(s))
	for i, v := range s {
		result[i] = f(v)
	}
	return result
}

func computeLayout(termW, termH int, cfg *config.Config) Layout {
	return Layout{
		ImageWidth:     int(float32(termW) * cfg.ImageWidthRatio),
		ImageHeight:    int(float32(termH) * cfg.ImageHeightRatio),
		SelectorWidth:  int(float32(termW) * cfg.SelectorWidthRatio),
		SelectorHeight: int(float32(termH) * cfg.SelectorHeightRatio),
		ChafaWidth:     int(float32(termW) * cfg.ImageWidthRatio),
		ChafaHeight:    int(float32(termH) * cfg.ImageHeightRatio),
		MaxFiles:       int(float32(termH) * cfg.SelectorHeightRatio),
	}
}

func constructFileList(m model) ([]string, int, int) {
	// fileNames
	names := Map(m.entries, func(e Entry) string {
		parts := strings.Split(e.name, "/")
		name := parts[len(parts)-1]
		runes := []rune(name)

		if len(runes) > 10 {
			return string(runes[:10])
		}
		return name
	})

	if len(m.entries) <= m.layout.MaxFiles {
		return names, m.cursor, 0
	}

	halfWay := m.layout.MaxFiles / 2

	if m.cursor < halfWay {
		return names[:m.layout.MaxFiles], m.cursor, 0
	}

	if m.cursor >= len(m.entries)-halfWay {
		cursorPos := m.cursor - (len(names) - m.layout.MaxFiles)
		return names[len(names)-m.layout.MaxFiles:], cursorPos, len(names) - m.layout.MaxFiles
	}

	top := m.cursor - halfWay
	bottom := m.cursor + halfWay
	return names[top:bottom], halfWay, top
}

func fileInfoString(imageInfo filesystem.ImageInfo) string {
	s := ""
	s += fmt.Sprintf("Name: %s\n", imageInfo.Name)
	s += fmt.Sprintf("File Size: %s %s\n", imageInfo.FileSize, imageInfo.FileSizeFormat)
	s += fmt.Sprintf("Dimensions %spx x %spx\n", imageInfo.Width, imageInfo.Height)
	return s
}

// Returns the full window range (all states)
func getWindowRange(m model) (int, int) {
	windowSize := 10
	start := m.cursor - windowSize/2
	end := start + windowSize
	if start < 0 {
		start = 0
	}
	if end > len(m.entries)-1 {
		end = len(m.entries) - 1
	}
	return start, end
}

func (m model) Init() tea.Cmd {
	return tea.ClearScreen
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case imageLoadedMsg:
		m.entries[msg.index].data = msg.image
		m.entries[msg.index].loadState = Loaded

	case imageUnloadedMsg:
		m.entries[msg.index].data = "NOT\nLOADED"
		m.entries[msg.index].loadState = Unloaded

	case tea.WindowSizeMsg:
		var cmds []tea.Cmd
		cmds = append(cmds, tea.ClearScreen)

		w, h := msg.Width, msg.Height
		m.layout = computeLayout(w, h, config.C)

		winStart, winEnd := getWindowRange(m)
		// recreate all chafa images
		// wipe all current chafa images
		for index, entry := range m.entries {
			entry.data = "NOT\nLOADED"
			entry.loadState = Unloaded
			if index >= winStart && index <= winEnd {
				cmds = append(cmds, loadImageCmd(m.chafaPath, m.layout, entry, index))
			}
		}
		return m, tea.Batch(cmds...)

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "Q":
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

			start, end := getWindowRange(m)
			var cmds []tea.Cmd

			for i := start; i <= end; i++ {
				if m.entries[i].loadState == Unloaded {
					m.entries[i].loadState = Loading
					cmds = append(cmds, loadImageCmd(m.chafaPath, m.layout, m.entries[i], i))
				}
			}

			// evict the entry that fell off the bottom of the window
			evict := end + 1
			if evict < len(m.entries) {
				m.entries[evict].data = "NOT\nLOADED"
				m.entries[evict].loadState = Unloaded
			}

			return m, tea.Batch(cmds...)

		case "down", "j":
			if m.cursor < len(m.entries)-1 {
				m.cursor++
			}

			start, end := getWindowRange(m)
			var cmds []tea.Cmd

			for i := start; i <= end; i++ {
				if m.entries[i].loadState == Unloaded {
					m.entries[i].loadState = Loading
					cmds = append(cmds, loadImageCmd(m.chafaPath, m.layout, m.entries[i], i))
				}
			}

			// evict the entry that fell off the top of the window
			evict := start - 1
			if evict >= 0 {
				m.entries[evict].data = "NOT\nLOADED"
				m.entries[evict].loadState = Unloaded
			}

			return m, tea.Batch(cmds...)

		case "C":
			m.clipboardManager.CopyImage(m.entries[m.cursor].name)

		case "c":
			m.clipboardManager.CopyString(m.entries[m.cursor].name)

		case "s":
			// possible future feature idk rn
			m.searchMode = true

		case "i":
			if m.infoMode {
				m.infoMode = false
			} else {
				m.infoMode = true
			}

		}
	}
	return m, nil
}

func (m model) View() string {
	// title := "\n Select an image to preview :)"
	fileList := ""

	truncFiles, relativeCursor, offset := constructFileList(m)
	for index, file := range truncFiles {
		realIndex := offset + index
		cursor := "  "
		fileName := file
		if relativeCursor == index {
			cursor = " >"

			if m.entries[realIndex].loadState == Loaded {
				fileName = styles.GreenBoldStyle().Render(fileName)
			} else {
				fileName = styles.BoldStyle().Render(fileName)
			}
		} else {
			if m.entries[realIndex].loadState == Loaded {
				fileName = styles.GreenStyle().Render(fileName)
			}
		}
		fileList += fmt.Sprintf("%s %s\n", cursor, fileName)
	}

	fileList = styles.SelectorBorderStyle(m.layout.SelectorWidth, m.layout.SelectorHeight).Render(fileList)
	if m.infoMode {
		imInfo, _ := filesystem.GetImageInfo(m.entries[m.cursor].name)
		imInfoString := fileInfoString(imInfo)
		fileList = styles.SelectorBorderStyle(m.layout.SelectorWidth, m.layout.SelectorHeight).Render(imInfoString)
	}

	rightSide := ""
	if m.cursor >= 0 && m.cursor < len(m.entries) {
		image := styles.ImageBorderStyle(m.layout.ImageWidth, m.layout.ImageHeight).Render(m.entries[m.cursor].data)
		caption := "(C)opy Image (c)opy image path (S)earch (Q)uit"

		rightSide = lipgloss.JoinVertical(
			lipgloss.Left,
			image,
			caption,
		)
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		fileList,
		"    ",
		rightSide,
	)
}

func main() {

	// imPath := "/home/olive/Pictures/stock-images/17564.png"
	// // getImageInfo(imPath)
	// imInfo, err := filesystem.GetImageInfo(imPath)
	// fmt.Println(imInfo)
	// return

	if err := config.Load("config/config.toml"); err != nil {
		fmt.Println(err.Error())
		return
	}

	clipboardManager := clipboard.GetClipboardManager()

	if err := clipboardManager.Setup(); err != nil {
		fmt.Println("clipboard setup failed:", err)
		return
	}

	// allow for no args and use current directory
	path, err := os.Getwd()

	if len(os.Args) > 2 {
		fmt.Println("Error: too many args")
		return
	} else if len(os.Args) == 2 {
		path = os.Args[1]
	}

	fmt.Println(path)

	if _, err := exec.LookPath("wl-copy"); err != nil {
		fmt.Fprintf(os.Stderr, "wl-copy not found: %v\n", err)
		os.Exit(1)
	}

	chafaPath, err := exec.LookPath("chafa")
	if err != nil {
		fmt.Println("chafa not found: %w", err)
	}
	fmt.Println(path)

	width, height, err := term.GetSize(int(os.Stdout.Fd()))

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	filePaths, err := imprev.GetFiles(path)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	entryList := make([]Entry, len(filePaths))

	for i := 0; i < len(filePaths); i++ {
		entryList[i].data = "NOT\nLOADED"
		entryList[i].loadState = Unloaded
		entryList[i].name = filePaths[i]
	}

	m := model{
		cursor:           0,
		layout:           computeLayout(width, height, config.C),
		entries:          entryList,
		chafaPath:        chafaPath,
		clipboardManager: clipboardManager,
	}

	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
