package clipboard

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"

	clip "golang.design/x/clipboard"
)

type ClipboardSetup func() (err error)

type CopyString func(s string) (err error)

type CopyImage func(s string) (err error)

type ClipboardManager struct {
	Setup      ClipboardSetup
	CopyString CopyString
	CopyImage  CopyImage
}

func getEnv() string {
	sessionType := os.Getenv("XDG_SESSION_TYPE")
	fmt.Println(sessionType)
	return sessionType
}

func waylandSetup() error {
	if _, err := exec.LookPath("wl-copy"); err != nil {
		return err
	}
	return nil
}

func nonWaylandSetup() error {
	err := clip.Init()
	if err != nil {
		return err
	}
	return nil
}

func waylandCopyString(s string) error {
	cmd := exec.Command("wl-copy", s)
	cmd.Run()
	if cmd.Err != nil {
		return cmd.Err
	}
	return nil
}

func waylandCopyImage(s string) error {
	// open image
	file, err := os.Open(s)
	if err != nil {
		// fmt.Println(err)
		return err
	}
	defer file.Close()

	// Get the file size
	stat, err := file.Stat()
	if err != nil {
		// fmt.Println(err)
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
	cmd := exec.Command("wl-copy", "--type", "image/png")
	cmd.Stdin = bytes.NewReader(bs)
	cmd.Run()
	return nil
}

func nonWaylandCopyString(s string) error {
	clip.Write(clip.FmtText, []byte(s))
	return nil
}

func nonWaylandCopyImage(s string) error {
	// open image
	file, err := os.Open(s)
	if err != nil {
		// fmt.Println(err)
		return err
	}
	defer file.Close()

	// Get the file size
	stat, err := file.Stat()
	if err != nil {
		// fmt.Println(err)
		return err
	}

	// Read the file into a byte slice
	bs := make([]byte, stat.Size())
	_, err = bufio.NewReader(file).Read(bs)
	if err != nil && err != io.EOF {
		fmt.Println(err)
		return err
	}
	clip.Write(clip.FmtImage, bs)
	return nil
}

func Test() {
	clip.Write(clip.FmtText, []byte("TEST"))
}

func GetClipboardManager() ClipboardManager {

	clipboardMgr := ClipboardManager{}

	if getEnv() == "wayland" {
		clipboardMgr.Setup = waylandSetup
		clipboardMgr.CopyImage = waylandCopyImage
		clipboardMgr.CopyString = waylandCopyString
	} else {
		clipboardMgr.Setup = nonWaylandSetup
		clipboardMgr.CopyString = nonWaylandCopyString
		clipboardMgr.CopyImage = nonWaylandCopyImage
	}
	return clipboardMgr
}
