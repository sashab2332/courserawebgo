package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

var prefix map[int]bool = map[int]bool{}

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


func dirTree(out io.Writer, path string, printFiles bool) error {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return err
	}

	sort.Slice(files, func(i, j int) bool { return files[i].Name() < files[j].Name() })

		lastDir := 0
		for i, el := range files {
			if el.IsDir() {
				lastDir=i
			}
		}

	for j, file := range files {
		filePath := filepath.FromSlash(filepath.Join(path, file.Name()))
		indent := len(strings.Split(filePath, string(os.PathSeparator))) - 2

		fileInfo, err := os.Stat(filePath)
		if err != nil {
			fmt.Println(err)
		}

		for i := 0; i < indent; i++ {
			if !printFiles && !fileInfo.IsDir() {
				break
			}
			if prefix[i] == false {
				fmt.Fprint(out, "│")
			}
			fmt.Fprint(out, "\t")
		}

		fileSize := strconv.FormatInt(file.Size(), 10)
		if file.Size() == 0 {
			fileSize = "(empty)"
		} else {
			fileSize = "(" + fileSize + "b)"
		}

		if !fileInfo.IsDir() {
			if printFiles {
				if j == len(files)-1 {
					fmt.Fprintf(out, "└───%v %v\n", file.Name(), fileSize)
					prefix[indent] = true
					//prefix = append(prefix, true)
				} else {
					fmt.Fprintf(out, "├───%v %v\n", file.Name(), fileSize)
					if prefix[indent] == true {
						prefix[indent] = false
					}
				}
			}
		} else {
			if j == len(files)-1 || (j ==lastDir && !printFiles){
				fmt.Fprintf(out, "└───%v\n", file.Name())
				prefix[indent] = true
			} else {
				fmt.Fprintf(out, "├───%v\n", file.Name())
				if prefix[indent] == true {
					prefix[indent] = false
				}
			}
		}
		dirTree(out, filePath, printFiles)
	}
	return nil
}
