package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	imprev "github.com/miabobia/im-prev"
	"github.com/miabobia/im-prev/config"
	"github.com/miabobia/im-prev/filesystem"
	"github.com/miabobia/im-prev/styles"
	"golang.org/x/term"
)

var renderers = map[Mode]func(model) string{
	FullScreen: renderFullscreen,
	Help:       renderHelpScreen,
	TooSmall:   renderScreenTooSmall,
	Info:       renderFileInfoScreen,
	Default:    renderDefault,
}

func renderFullscreen(m model) string {

	physicalWidth, _, _ := term.GetSize(int(os.Stdout.Fd()))

	chafaImage, _ := imprev.GetChafaImage(
		config.C.ChafaDefaultSymbols,
		fmt.Sprintf("%dx%d", m.layout.ImageFullscreenWidth, m.layout.ImageFullscreenHeight),
		m.chafaPath,
		m.entries[m.cursor].name,
	)
	content := styles.ImageBorderStyle(m.layout.ImageFullscreenWidth, m.layout.ImageFullscreenHeight).
		AlignHorizontal(lipgloss.Center).
		Render(chafaImage)

	centeredContent := lipgloss.PlaceHorizontal(
		physicalWidth,
		lipgloss.Center,
		content,
	)

	return centeredContent

}

func renderScreenTooSmall(m model) string {
	return "WINDOW TOO SMALL!"
}

func renderFileInfoScreen(m model) string {

	fileList := ""
	fileList += fmt.Sprintf("%s\n", m.cwd)

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
	// if m.mode == Info {
	imInfo, _ := filesystem.GetImageInfo(m.entries[m.cursor].name)
	imInfoString := fileInfoString(imInfo)
	fileList = styles.SelectorBorderStyle(m.layout.SelectorWidth, m.layout.SelectorHeight).Render(imInfoString)
	// }

	rightSide := ""
	if m.cursor >= 0 && m.cursor < len(m.entries) {
		image := styles.ImageBorderStyle(m.layout.ImageWidth, m.layout.ImageHeight).Render(m.entries[m.cursor].data)
		caption := "(?) Help"

		rightSide = lipgloss.JoinVertical(
			lipgloss.Left,
			image,
			caption,
		)
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		fileList,
		// "    ",
		rightSide,
	)
}

func renderDefault(m model) string {

	fileList := ""
	fileList += fmt.Sprintf("%s\n", m.cwd)

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

	rightSide := ""
	if m.cursor >= 0 && m.cursor < len(m.entries) {
		image := styles.ImageBorderStyle(m.layout.ImageWidth, m.layout.ImageHeight).Render(m.entries[m.cursor].data)
		caption := "(?) Help"

		rightSide = lipgloss.JoinVertical(
			lipgloss.Left,
			image,
			caption,
		)
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		fileList,
		// "    ",
		rightSide,
	)
}

func renderHelpScreen(m model) string {

	helperString := "[C] -> Copy image data\n[P] -> Copy image     \n[I] -> Info mode      \n[F] -> Fullscreen mode\n[Q] -> Quit           "

	physicalWidth, physicalHeight, _ := term.GetSize(int(os.Stdout.Fd()))

	content := styles.ImageBorderStyle(m.layout.ImageFullscreenWidth, m.layout.ImageFullscreenHeight).
		AlignHorizontal(lipgloss.Center).
		Render(helperString)

	centeredContent := lipgloss.PlaceHorizontal(
		physicalWidth,
		lipgloss.Center,
		content,
	)

	// lipgloss.Place()
	helperString = styles.SelectorBorderStyle(24, 7).Render(helperString)
	return lipgloss.Place(physicalWidth, physicalHeight, lipgloss.Center, lipgloss.Center, helperString)

	return centeredContent

}
