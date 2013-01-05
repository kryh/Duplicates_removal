package main

import (
	"fmt"
	"os"
	"path/filepath"
	"io/ioutil"
	"crypto/sha1"
	"runtime" //GOMAXPROCS(NumCPU())
	"sync"
)

var booksMap = make(map[int64][]string)
var done chan bool

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

func handleTableOfBooksWithTheSameSize(listOfBooks []string, mapHashFile map[string]string, mutex *sync.Mutex) {
	numberOfFilesWithTheSameSize := len(listOfBooks)
	
	//create enough channels to handle checksum calculation for all files with particular size
	channels := make([]chan string, numberOfFilesWithTheSameSize)
	
	for i:=0; i<numberOfFilesWithTheSameSize; i++ {
		channels[i] = make(chan string)
		
		go calculateChecksum(listOfBooks[i], channels[i])
	}
		
	for i:=0; i<numberOfFilesWithTheSameSize; i++ {
		bookname := listOfBooks[i]
		calculatedHash := <-channels[i]
		
		mutex.Lock()
		if _,ok := mapHashFile[calculatedHash]; !ok {  //check if hash already exists
			mapHashFile[calculatedHash] = bookname
		} else {
			fmt.Println("Remove:", bookname)
			os.Remove(bookname)
		}
		mutex.Unlock()
	}
	done <- true
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	addBooksToMap("_Przeczytane")
	addBooksToMap("_Przeczytane_do_przejrzenia")
	addBooksToMap("Sorted")
	addBooksToMap("Unsorted")

	mapHashFilename := make(map[string]string)
	mutex := new(sync.Mutex)
	
	numberOfGoroutinesToWaitFor := 0
	for _, names := range booksMap {
		if len(names) > 1 {
			numberOfGoroutinesToWaitFor++;
		}
	}
	
	done = make(chan bool, numberOfGoroutinesToWaitFor)
	
	for _, names := range booksMap {
		if len(names) > 1 {
			go handleTableOfBooksWithTheSameSize(names, mapHashFilename, mutex)
		}
	}
	
	//wait until all goroutines finish
	for i:=0; i<numberOfGoroutinesToWaitFor; i++ {
		<- done
	}
}
