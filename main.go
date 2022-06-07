package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"strings"

	"github.com/awesome-gocui/gocui"
)

var (
	viewArr          = []string{"First", "Second"}
	active           = 0
	viewDir          = []string{"", ""}
	dirIcon          = "\uf413"
	fileIcon         = "\uf016"
	numberOfElements = []int{0, 0}
	numberOfDown     = []int{0, 0}
)

func verifyFormatting(str string) string {
	if str[0] != '/' {
		str = "/" + str
	}

	return str
}

func isDirectory(path string) (bool, error) {
	file, errOs := os.Stat(path)
	if errOs != nil {
		return false, errOs
	}

	if !file.IsDir() {
		return false, nil
	}

	return true, nil
}

func refreshAllViews(g *gocui.Gui) error {
	for act, name := range viewArr {
		if v, err := g.View(name); err != nil {
			return err
		} else {
			renderDir(g, v, act)
		}
	}

	return nil
}

func goBack(path string) string {
	str := path[:strings.LastIndex(path, "/")]
	if len(str) < 1 {
		return "/"
	}
	return str
}

func appendPath(curr, add string) string {
	if curr[len(curr)-1] == '/' {
		return curr + add
	}
	return curr + "/" + add
}

func inputChangeDir(g *gocui.Gui, _ *gocui.View) error {
	maxX, maxY := g.Size()

	if input, err := g.SetView("Input", maxX/2-20, maxY/2-1, maxX/2+20, maxY/2+1, 0); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		input.Editable = true

		if _, err = setCurrentViewOnTop(g, "Input"); err != nil {
			return err
		}
	}

	if err := g.SetKeybinding("Input", gocui.KeyEnter, gocui.ModNone,
		func(g *gocui.Gui, v *gocui.View) error {
			focusedView, err := g.SetCurrentView(viewArr[active])
			if err != nil {
				return err
			}
			result := v.ViewBuffer()
			result = strings.TrimSpace(result)

			if err := g.DeleteView("Input"); err != nil {
				return err
			}

			if _, err := setCurrentViewOnTop(g, viewArr[active]); err != nil {
				return err
			}

			viewDir[active] = verifyFormatting(result)
			renderDir(g, focusedView, active)

			return nil
		}); err != nil {
		log.Panicln("input create:", err)
	}

	return nil
}

func setCurrentViewOnTop(g *gocui.Gui, name string) (*gocui.View, error) {
	if _, err := g.SetCurrentView(name); err != nil {
		return nil, err
	}
	return g.SetViewOnTop(name)
}

func nextView(g *gocui.Gui, v *gocui.View) error {
	nextIndex := (active + 1) % len(viewArr)
	name := viewArr[nextIndex]

	new, err := setCurrentViewOnTop(g, name)
	if err != nil {
		return err
	}

	v.Highlight = false
	new.Highlight = true

	active = nextIndex
	return nil
}

func renderDir(g *gocui.Gui, v *gocui.View, index int) error {
	files, err := ioutil.ReadDir(viewDir[index])
	if err != nil {
		v.Clear()
		v.Title = viewDir[index]
		fmt.Fprintf(v, "Error: path not found")
		return nil
	}

	v.Clear()
	v.Title = viewDir[index]

	numberOfElements[index] = 0
	numberOfDown[index] = 0
	fmt.Fprint(v, "..\n")
	numberOfElements[index]++
	for _, file := range files {
		numberOfElements[index]++
		if file.IsDir() {
			fmt.Fprintf(v, "%v %v\n", dirIcon, file.Name())
			continue
		}

		fmt.Fprintf(v, "%v %v\n", fileIcon, file.Name())
	}

	return nil
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView("First", 0, 0, maxX/2-1, maxY-2, 0); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Highlight = true

		renderDir(g, v, 0)

		if _, err = setCurrentViewOnTop(g, "First"); err != nil {
			return err
		}
	}

	if v, err := g.SetView("Second", maxX/2, 0, maxX-1, maxY-2, 0); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Highlight = false

		renderDir(g, v, 1)
	}

	if v, err := g.SetView("Third", -1, maxY-2, maxX, maxY, 1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Highlight = false
		v.Frame = false
		v.FgColor = gocui.ColorBlack
		v.BgColor = gocui.ColorGreen
		fmt.Fprintf(v, "Tab - switch view | Space - cd | ArrowLeft/ArrowRight - Move file/dir")
	}

	return nil
}

func cursorDown(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		cx, cy := v.Cursor()
		ox, oy := v.Origin()
		_, maxY := v.Size()

		// Prevent infinite scrolling
		if numberOfElements[active] < numberOfDown[active]+2 {
			return nil
		}

		numberOfDown[active]++

		if maxY < cy+2 {
			if err := v.SetOrigin(ox, oy+1); err != nil {
				return err
			}
			return nil
		}

		if err := v.SetCursor(cx, cy+1); err != nil {
			return err
		}
	}
	return nil
}

func cursorUp(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		ox, oy := v.Origin()
		cx, cy := v.Cursor()
		if oy == 0 && cy == 0 {
			return nil
		}

		if err := v.SetCursor(cx, cy-1); err != nil && oy > 0 {
			if err := v.SetOrigin(ox, oy-1); err != nil {
				return err
			}
		}

		numberOfDown[active]--
	}
	return nil
}

func openDir(g *gocui.Gui, v *gocui.View) error {
	var l string
	var err error

	_, cy := v.Cursor()
	if l, err = v.Line(cy); err != nil {
		l = ""
	}

	l = strings.Replace(l, dirIcon, "", -1)
	l = strings.Replace(l, fileIcon, "", -1)
	l = strings.TrimSpace(l)

	if l == ".." {
		goBackController(g, v)
		return nil
	}

	path := appendPath(viewDir[active], l)

	isDir, errOs := isDirectory(path)
	if errOs != nil {
		return errOs
	}

	if !isDir {
		return nil
	}

	viewDir[active] = path
	renderDir(g, v, active)
	return nil
}

func goBackController(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		viewDir[active] = goBack(viewDir[active])
		renderDir(g, v, active)
	}
	return nil
}

func moveFile(g *gocui.Gui, v *gocui.View) error {
	if viewDir[0] == viewDir[1] {
		return nil
	}

	var l string
	var err error

	_, cy := v.Cursor()
	if l, err = v.Line(cy); err != nil {
		l = ""
	}

	l = strings.Replace(l, dirIcon, "", -1)
	l = strings.Replace(l, fileIcon, "", -1)
	l = strings.TrimSpace(l)

	if l == ".." {
		return nil
	}

	if v.Name() == viewArr[0] {
		if errMv := os.Rename(appendPath(viewDir[0], l), appendPath(viewDir[1], l)); errMv != nil {
			return errMv
		}
	} else {
		if errMv := os.Rename(appendPath(viewDir[1], l), appendPath(viewDir[0], l)); errMv != nil {
			return errMv
		}
	}

	refreshAllViews(g)
	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func main() {
	if runtime.GOOS == "windows" {
		fmt.Println("Sorry, this program is not supported on Windows")
		os.Exit(0)
	}

	if len(os.Args[1:]) != 2 {
		fmt.Printf("Usage: %s <dir1> <dir2>\n", os.Args[0])
		os.Exit(0)
	}
	viewDir[0] = verifyFormatting(os.Args[1])
	viewDir[1] = verifyFormatting(os.Args[2])

	g, err := gocui.NewGui(gocui.OutputNormal, true)
	if err != nil {
		log.Panicln("newgui:", err)
	}
	defer g.Close()

	g.Highlight = true
	g.SelFgColor = gocui.ColorGreen
	g.SelFrameColor = gocui.ColorGreen

	g.SetManagerFunc(layout)

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln("keybinding:", err)
	}

	if err := g.SetKeybinding("", gocui.KeyTab, gocui.ModNone, nextView); err != nil {
		log.Panicln("keybinding:", err)
	}

	if err := g.SetKeybinding("", gocui.KeySpace, gocui.ModNone, inputChangeDir); err != nil {
		log.Panicln("keybinding:", err)
	}

	if err := g.SetKeybinding("", gocui.KeyArrowUp, gocui.ModNone, cursorUp); err != nil {
		log.Panicln("keybinding:", err)
	}

	if err := g.SetKeybinding("", gocui.KeyArrowDown, gocui.ModNone, cursorDown); err != nil {
		log.Panicln("keybinding:", err)
	}

	if err := g.SetKeybinding("First", gocui.KeyEnter, gocui.ModNone, openDir); err != nil {
		log.Panicln("keybinding:", err)
	}

	if err := g.SetKeybinding("Second", gocui.KeyEnter, gocui.ModNone, openDir); err != nil {
		log.Panicln("keybinding:", err)
	}

	if err := g.SetKeybinding("First", gocui.KeyBackspace2, gocui.ModNone, goBackController); err != nil {
		log.Panicln("keybinding:", err)
	}

	if err := g.SetKeybinding("Second", gocui.KeyBackspace2, gocui.ModNone, goBackController); err != nil {
		log.Panicln("keybinding:", err)
	}

	if err := g.SetKeybinding("First", gocui.KeyArrowRight, gocui.ModNone, moveFile); err != nil {
		log.Panicln("keybinding:", err)
	}

	if err := g.SetKeybinding("Second", gocui.KeyArrowLeft, gocui.ModNone, moveFile); err != nil {
		log.Panicln("keybinding:", err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln("mainloop:", err)
	}
}
