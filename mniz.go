/*
MIT License

Copyright (c) 2022 a5traa

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

// rewrite disassm with https://github.com/knightsc/gapstone instead of using native disassm
// check https://www.capstone-engine.org/ for better refrence
// this program will be documented soon

package main

import (
	"encoding/hex"
	"fmt"
	"os"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

const chunk = 1024

func main() {
	a := app.New()
	w := a.NewWindow("mniz")
	w.Resize(fyne.NewSize(800, 600))

	icon, err := os.ReadFile("assets/picture.png") 
	if err == nil {
		w.SetIcon(fyne.NewStaticResource("picture.png", icon))
	}

	var path string
	var data []byte
	var offset int
	hexLbl := widget.NewLabel("")
	fileLbl := widget.NewLabel("not loaded")
	hexScrl := container.NewScroll(hexLbl)
	hexScrl.SetMinSize(fyne.NewSize(0, 300))

	openBtn := widget.NewButton("Open", func() {
		dialog.ShowFileOpen(func(r fyne.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			if r == nil {
				dialog.ShowInformation("error", "no file selected", w)
				return
			}

			path = r.URI().Path()
			name := r.URI().Name()
			go func() {
				data, err = os.ReadFile(path)
				if err != nil {
					fyne.CurrentApp().SendNotification(&fyne.Notification{
						Title:   "error",
						Content: fmt.Sprintf("failed to read file: %v", err),
					})
					return
				}
				if len(data) == 0 {
					dialog.ShowInformation("error", "file empty", w)
					return
				}
				updateHex(data, offset, hexLbl)
				fileLbl.SetText("file: " + name)
			}()
		}, w)
	})

	offsetEnt := widget.NewEntry()
	valEnt := widget.NewEntry()

	modBtn := widget.NewButton("modify byte", func() {
		if len(data) == 0 {
			dialog.ShowInformation("error", "no file loaded", w)
			return
		}

		off, err := strconv.ParseInt(offsetEnt.Text, 16, 64)
		if err != nil || off < 0 || off >= int64(len(data)) {
			dialog.ShowError(fmt.Errorf("invalid?"), w)
			return
		}

		val, err := strconv.ParseUint(valEnt.Text, 16, 8)
		if err != nil {
			dialog.ShowError(fmt.Errorf("invalid?"), w)
			return
		}

		data[off] = byte(val)
		updateHex(data, offset, hexLbl)
		dialog.ShowInformation("success", "byte modified", w)
	})

	saveBtn := widget.NewButton("save File", func() {
		if len(data) == 0 || path == "" {
			dialog.ShowInformation("error", "no file loaded", w)
			return
		}
		err := os.WriteFile(path, data, os.ModePerm)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		dialog.ShowInformation("sucess", "file saved", w)
	})

	prevBtn := widget.NewButton("prev chunk", func() {
		if offset > 0 {
			offset -= chunk
			updateHex(data, offset, hexLbl)
		}
	})

	nextBtn := widget.NewButton("next chunk", func() {
		if offset+chunk < len(data) {
			offset += chunk
			updateHex(data, offset, hexLbl)
		}
	})

	disasmTbl := widget.NewTable(
		func() (int, int) { return (len(data)+chunk-1)/chunk + 1, 4 },
		func() fyne.CanvasObject {
			return container.NewVBox(widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{}))
		},
		func(id widget.TableCellID, o fyne.CanvasObject) {
			lbl := o.(*fyne.Container).Objects[0].(*widget.Label)
			if id.Row == 0 {
				switch id.Col {
				case 0:
					lbl.SetText("chunk start")
				case 1:
					lbl.SetText("chunk end")
				case 2:
					lbl.SetText("dex data")
				case 3:
					lbl.SetText("disassembly")
				}
				return
			}

			chunkStart := (id.Row - 1) * chunk
			if chunkStart >= len(data) {
				lbl.SetText("")
				return
			}
			chunkEnd := chunkStart + chunk
			if chunkEnd > len(data) {
				chunkEnd = len(data)
			}
			d := data[chunkStart:chunkEnd]
			switch id.Col {
			case 0:
				lbl.SetText(fmt.Sprintf("%-15s", fmt.Sprintf("0x%08X", chunkStart)))
			case 1:
				lbl.SetText(fmt.Sprintf("%-15s", fmt.Sprintf("0x%08X", chunkEnd)))
			case 2:
				lbl.SetText(fmt.Sprintf("%-60s", hex.EncodeToString(d)))
			case 3:
				disasm := disasm(d)
				lbl.SetText(fmt.Sprintf("disassembly: %-40s", disasm))
			}
		},
	)

	disasmTbl.SetColumnWidth(0, 150)
	disasmTbl.SetColumnWidth(1, 150)
	disasmTbl.SetColumnWidth(2, 300)
	disasmTbl.SetColumnWidth(3, 200)

	codeBox := widget.NewMultiLineEntry()
	codeBox.SetPlaceHolder("write code or modify here")
	codeBox.Wrapping = fyne.TextWrapWord

	disasmScrl := container.NewScroll(disasmTbl)

	tabs := container.NewAppTabs(
		container.NewTabItem("hex viewer", container.NewVBox(
			container.NewHBox(openBtn, saveBtn),
			container.NewVBox(
				widget.NewLabel("modify byte:"),
				container.NewHBox(widget.NewLabel("offset:"), offsetEnt, widget.NewLabel("New Value:"), valEnt, modBtn),
			),
			fileLbl,
			hexScrl,
			container.NewHBox(prevBtn, nextBtn),
		)),
		container.NewTabItem("disassembler", disasmScrl),
		container.NewTabItem("data processing", widget.NewLabel("data processing tab - to be written soon")),
		container.NewTabItem("code edit", codeBox),
	)

	w.SetContent(tabs)
	w.ShowAndRun()
}

func updateHex(data []byte, offset int, hexLbl *widget.Label) {
	end := offset + chunk
	if end > len(data) {
		end = len(data)
	}
	hexLbl.SetText(formatHex(data[offset:end]))
}

func formatHex(data []byte) string {
	const rowSize = 16
	rows := ""
	for i := 0; i < len(data); i += rowSize {
		end := i + rowSize
		if end > len(data) {
			end = len(data)
		}
		hexPart := hex.EncodeToString(data[i:end])
		asciiPart := ""
		for _, b := range data[i:end] {
			if b >= 32 && b <= 126 {
				asciiPart += string(b)
			} else {
				asciiPart += "."
			}
		}
		rows += fmt.Sprintf("%08X  %-47s  %s\n", i, hexPart, asciiPart)
	}
	return rows
}

func disasm(data []byte) string {
	if len(data) == 0 {
		return "empty Data"
	}

	op := data[0]

	if op == 0x90 {
		return "NOP"
	}

	if op == 0x8B && len(data) > 1 {
		return fmt.Sprintf("MOV %s, %s", regName(data[1]), regName(data[2]))
	}

	if op == 0x03 && len(data) > 1 {
		return fmt.Sprintf("ADD %s, %s", regName(data[1]), regName(data[2]))
	}

	if op == 0x2B && len(data) > 1 {
		return fmt.Sprintf("SUB %s, %s", regName(data[1]), regName(data[2]))
	}

	if op == 0xE9 && len(data) > 1 {
		off := int(int16(data[1]) | int16(data[2])<<8)
		return fmt.Sprintf("JMP 0x%04X", off)
	}

	return fmt.Sprintf("unknown instruction: 0x%02X", op)
}

func regName(r byte) string {
	regs := []string{
		"AL", "CL", "DL", "BL", "AH", "CH", "DH", "BH", "AX", "CX", "DX", "BX", "SP", "BP", "SI", "DI", "EAX", "ECX", "EDX", "EBX", "ESP", "EBP", "ESI", "EDI",
	}
	if int(r) < len(regs) {
		return regs[r]
	}
	return fmt.Sprintf("unknown reg: 0x%02X", r)
}
