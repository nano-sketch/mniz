# mniz - A Simple Hex Viewer

**mniz** is a lightweight and easy to use hex viewer built with golang and the fyne framework. Allows you to inspect the contents of files in hexadecimal format, making it useful for debugging or analyzing binary files.

![item2](https://github.com/user-attachments/assets/e25644b2-5160-44c0-be7a-112953722a21)


---

## Features
- **file loading**: Open and view any file in hexadecimal format.
- **scrollable view**: Easily navigate through the file's contents.
- **live file information**: Displays the current file's name and status.
- **chunking**: An efficient method for optimisation loading the file in 1024 chunks per page
---

## Installation
1. Clone the repository:
   ```bash
   git clone https://github.com/nano-sketch/mniz.git
   cd mniz
   ```
2. For windows & Unix
   ```bash
   go mod vendor
   go run .
   ```
   

 # Requirements
   ```bash
   go get fyne.io/fyne/v2
   ```
 You also need go v1.19 or higher

 ### you may contribute to my project at your own will, note: this project is incomplete and it is recommended to write the disassm using capstone framework https://github.com/knightsc/gapstone as writing a native disassm would be an issue. The code edit is also for loading the disassembled file but i could not implement yet. Will be adding future updates..
   
