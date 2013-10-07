package main

import (
	"crypto/sha1"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime" //GOMAXPROCS(NumCPU())
	"sync"
)

var booksMap = make(map[int64][]string)
var wg = new(sync.WaitGroup)


const MAXNUMBEROFOPENFILES = 80
var openFilesLimit = make(chan bool, MAXNUMBEROFOPENFILES)

//addBook adds a book file to map
func addBook(path string, info os.FileInfo, err error) error {
	if !info.IsDir() {
		booksMap[info.Size()] = append(booksMap[info.Size()], path)
	}
	return err
}

func addBooksToMap(folder string) {
	filepath.Walk(folder, addBook)
}

//calculateChecksum gets checksum of a file
func calculateChecksum(filename string, channel chan string, openFilesLimit chan bool) {
	openFilesLimit <- true
	defer func() {
		<- openFilesLimit
	}()

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

	for i := 0; i < numberOfFilesWithTheSameSize; i++ {
		channels[i] = make(chan string)

		go calculateChecksum(listOfBooks[i], channels[i], openFilesLimit)
	}

	for i := 0; i < numberOfFilesWithTheSameSize; i++ {
		bookname := listOfBooks[i]
		calculatedHash := <-channels[i]

		mutex.Lock()
		if _, ok := mapHashFile[calculatedHash]; !ok { //check if hash already exists
			mapHashFile[calculatedHash] = bookname
		} else {
			fmt.Println("Remove:", bookname)
 			os.Remove(bookname)
		}
		mutex.Unlock()
	}

	wg.Done()
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	addBooksToMap("_Przeczytane")
	addBooksToMap("Sorted")
	addBooksToMap("Unsorted")

	mapHashFilename := make(map[string]string)
	mutex := new(sync.Mutex)
	
	for _, names := range booksMap {
		if len(names) > 1 {
			wg.Add(1)
			go handleTableOfBooksWithTheSameSize(names, mapHashFilename, mutex)
		}
	}

	wg.Wait()
}
