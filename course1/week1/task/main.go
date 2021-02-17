package main

import (
	"io"
	"io/ioutil"
	"os"
	"strconv"
)

func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"

	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
}

func dirTree(out io.StringWriter, path string, printFiles bool) (err error) {
	return subDirTree(out, path, printFiles, "")
}

func subDirTree(out io.StringWriter, path string, printFiles bool, level string) (err error) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return err
	}

	files = filterFiles(files, printFiles)

	for n, file := range files {
		out.WriteString(level)

		lastRow := n == len(files)-1

		newLevel := level
		if lastRow {
			out.WriteString("└───")
			newLevel += "\t"
		} else {
			out.WriteString("├───")
			newLevel += "│\t"
		}

		out.WriteString(file.Name())
		if file.IsDir() {
			out.WriteString("\n")
			err = subDirTree(out, path+string(os.PathSeparator)+file.Name(), printFiles, newLevel)
			if err != nil {
				return err
			}
		} else {

			size := strconv.FormatInt(file.Size(), 10) + "b"
			if size == "0b" {
				size = "empty"
			}

			out.WriteString(" (" + size + ")\n")
		}

	}
	return nil
}

func filterFiles(files []os.FileInfo, printFiles bool) []os.FileInfo {
	newFiles := make([]os.FileInfo, 0, cap(files))
	for _, file := range files {
		if ".DS_Store" == file.Name() {
			continue
		}
		if !file.IsDir() && !printFiles {
			continue
		}
		newFiles = append(newFiles, file)
	}
	return newFiles
}
