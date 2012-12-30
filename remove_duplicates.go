package main

import (
	"fmt"
	"os"
	"path/filepath"
	"io/ioutil"
	"crypto/sha1"
	"runtime" //GOMAXPROCS(NumCPU())
)

var booksMap = make(map[int64][]string)


func addBook(path string, info os.FileInfo, err error) error {
	if !info.IsDir() {
		booksMap[info.Size()] = append(booksMap[info.Size()], path)
	}
	return err
}

func addBooksToMap(folder string) {
	filepath.Walk(folder, addBook)
}

func calculateChecksum(filename string, channel chan string) {
	fmt.Println("Calc sha1 for: " + filename)
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println("Error during opening " + filename)
		return
	}
	h := sha1.New()
	h.Write(data)
	channel <- fmt.Sprintf("%x", h.Sum(nil))
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	addBooksToMap("_Przeczytane")
	addBooksToMap("_Przeczytane_do_przejrzenia")
	addBooksToMap("Sorted")
	addBooksToMap("Unsorted")

 	mapHashFilename := make(map[string]string)
	
	for _, names := range booksMap {
		numberOfFilesWithTheSameSize := len(names)
		
		if numberOfFilesWithTheSameSize > 1 {
			
			//create enough channels to handle checksum calculation for all files with particular size
			channels := make([]chan string, numberOfFilesWithTheSameSize)
			
			for i:=0; i<numberOfFilesWithTheSameSize; i++ {
				channels[i] = make(chan string)
				
				go calculateChecksum(names[i], channels[i])
			}
			
			for i:=0; i<numberOfFilesWithTheSameSize; i++ {
				bookname := names[i]
				calculatedHash := <-channels[i]
				
 				if _,ok := mapHashFilename[calculatedHash]; !ok {  //check if hash already exists
 					mapHashFilename[calculatedHash] = bookname
 				} else {
 					fmt.Println("Removing book: ", bookname, "\n")
					os.Remove(bookname)
 				}
			}
 		}
	}
}
