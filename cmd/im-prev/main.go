// =TODO=
// search bar
// directory jumping ?
// recursive directory search
// help
// version
// vertical size inconsistent

// =NEXT=
// filename truncating
// current directory

// =DONE=
// file support for webp, bmp,
// info section
// fullscreen
// min window size indicator

package main

import (
	"fmt"
	"os"
	"os/exec"

	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/term"

	// local packages
	imprev "github.com/miabobia/im-prev"
	"github.com/miabobia/im-prev/clipboard"
	"github.com/miabobia/im-prev/config"
	"github.com/miabobia/im-prev/filesystem"
)

type Mode int

const (
	Default Mode = iota
	FullScreen
	Info
	Help
	TooSmall
)

type model struct {
	entries          []Entry
	cursor           int
	layout           Layout
	chafaPath        string
	clipboardManager clipboard.ClipboardManager
	cwd              string
	mode             Mode
}

type Layout struct {
	ImageWidth            int
	ImageHeight           int
	SelectorWidth         int
	SelectorHeight        int
	ChafaWidth            int
	ChafaHeight           int
	MaxFiles              int
	ImageFullscreenWidth  int
	ImageFullscreenHeight int
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

type ScrollingText struct {
	scrolling bool
	text      string
	n         int
}

// Tea messages
type imageLoadedMsg struct {
	index int
	image string
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
		ImageWidth:            int(float32(termW) * cfg.ImageWidthRatio),
		ImageHeight:           int(float32(termH) * cfg.ImageHeightRatio),
		SelectorWidth:         int(float32(termW) * cfg.SelectorWidthRatio),
		SelectorHeight:        int(float32(termH) * cfg.SelectorHeightRatio),
		ChafaWidth:            int(float32(termW) * cfg.ImageWidthRatio),
		ChafaHeight:           int(float32(termH) * cfg.ImageHeightRatio),
		MaxFiles:              int(float32(termH) * cfg.SelectorHeightRatio),
		ImageFullscreenWidth:  int(float32(termW) * cfg.ImageFullscreenWidthRatio),
		ImageFullscreenHeight: int(float32(termH) * cfg.ImageFullscreenHeightRatio),
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
	s += fmt.Sprintf("Last Modified: %s", imageInfo.ModifiedTime)
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

	case tea.WindowSizeMsg:
		return handleWindowResize(m, msg)
		// var cmds []tea.Cmd
		// cmds = append(cmds, tea.ClearScreen)

		// w, h := msg.Width, msg.Height
		// m.layout = computeLayout(w, h, config.C)

		// if w <= config.C.MinimumScreenWidth || h <= config.C.MinimumScreenHeight {
		// 	m.mode = TooSmall
		// }

		// winStart, winEnd := getWindowRange(m)
		// // recreate all chafa images
		// // wipe all current chafa images
		// for index, entry := range m.entries {
		// 	entry.data = "NOT\nLOADED"
		// 	entry.loadState = Unloaded
		// 	if index >= winStart && index <= winEnd {
		// 		cmds = append(cmds, loadImageCmd(m.chafaPath, m.layout, entry, index))
		// 	}
		// }
		// return m, tea.Batch(cmds...)

	case tea.KeyMsg:
		return handleKeyPress(m, msg)

	}
	return m, nil
}

func (m model) View() string {
	return renderers[m.mode](m)
}

func main() {

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
		cwd:              path,
		mode:             Default,
	}

	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
