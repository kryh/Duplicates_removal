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

type response struct {
	position int
	hash string
}

type request struct {
	position int
	filename string
	respchan chan response
}

const START_RED = "\x1b[31;1m"
const START_GREEN = "\x1b[32;1m"
const END_COLOR = "\x1b[0m"

const MAXNUMBEROFOPENFILES = 40
var openFilesLimit = make(chan bool, MAXNUMBEROFOPENFILES)

var REDUCED_SIZE int64 = 0


//addBook adds a book file to map
func addBook(path string, info os.FileInfo, err error) error {
	if info.Mode().IsRegular() {	//handle only regular files, no folders, symlinks etc
		booksMap[info.Size()] = append(booksMap[info.Size()], path)
	}
	return err
}

func addBooksToMap(folder string) {
	filepath.Walk(folder, addBook)
}

//calculateChecksum gets checksum of a file
func calculateChecksum(req *request) {
 	fmt.Println("Calc sha1 for: " + req.filename)
	data, err := ioutil.ReadFile(req.filename)
	if err != nil {
		fmt.Println("Error during opening " + req.filename)
		return
	}
	hash := sha1.New()
	hash.Write(data)
	
	req.respchan <- response{position:req.position, hash:fmt.Sprintf("%x", hash.Sum(nil))}
}

func calculateChecksum2(req *request) {
	fmt.Println("Calc2 sha1 for: " + req.filename)
	file, err := os.Open(req.filename)
	if err != nil {
		fmt.Println("Error during opening " + req.filename)
		return
	}
	defer file.Close()
	
	const bufferSize = 30*1024*1024
	buffer := make([]byte, bufferSize)
	hash := sha1.New()
	for {
		readLength, err := file.Read(buffer)
		if err != nil {
			panic(err)
		}
		hash.Write(buffer[:readLength])
		
		if readLength < bufferSize {
			break
		}
	}
	
	req.respchan <- response{position:req.position, hash:fmt.Sprintf("%x", hash.Sum(nil))}
}


func handleTableOfBooksWithTheSameSize(bookSize int64, listOfBooks []string, mapHashFile map[string]string, mutex *sync.Mutex) {
	numberOfFilesWithTheSameSize := len(listOfBooks)

	//create enough channels to handle checksum calculation for all files with particular size
	channel := make(chan response, numberOfFilesWithTheSameSize)

	for i := 0; i < numberOfFilesWithTheSameSize; i++ {
		i := i
		/*
		 * In a Go for loop, the loop variable is reused for each iteration, so the i variable is shared across all goroutines.
		 * That's not what we want. We need to make sure that i is unique for each goroutine.  
		 */
		openFilesLimit <- true
		go func() {
			calculateChecksum2(&request{position:i, filename:listOfBooks[i], respchan:channel})
			<- openFilesLimit
			runtime.GC()	//run garbace collector to clean up buffer memory
		}()		
	}

	sliceOfHashStrings := make([]string, numberOfFilesWithTheSameSize)
	
	for i := 0; i< numberOfFilesWithTheSameSize; i++ {
		tempResp := <- channel
		sliceOfHashStrings[tempResp.position] = tempResp.hash
	}
	
	for i := 0; i < numberOfFilesWithTheSameSize; i++ {
		bookname := listOfBooks[i]
		calculatedHash := sliceOfHashStrings[i]

		mutex.Lock()
		if _, ok := mapHashFile[calculatedHash]; !ok { //check if hash already exists
			mapHashFile[calculatedHash] = bookname
		} else {
			
			fmt.Println(START_RED+"Removing:", bookname, END_COLOR)
 			os.Remove(bookname)
			REDUCED_SIZE += bookSize
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
	
	for bookSize, names := range booksMap {
		if len(names) > 1 {
			wg.Add(1)
			go handleTableOfBooksWithTheSameSize(bookSize, names, mapHashFilename, mutex)
		}
	}

	wg.Wait()
	
	fmt.Printf(START_GREEN+"Removed: %d MB.\n"+END_COLOR, REDUCED_SIZE/1024/1024)
}
