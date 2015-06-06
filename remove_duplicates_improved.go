package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"crypto/sha1"
)

var wg sync.WaitGroup

const (
	START_RED = "\x1b[31;1m"
	START_GREEN = "\x1b[32;1m"
	END_COLOR = "\x1b[0m"

	OPEN_FILES_LIMIT = 50
)

var openFileLimitChan = make(chan bool, OPEN_FILES_LIMIT)


func populateMap(booksMap map[int64][]string, folderSlice []string) {
	//walk through each of 3 folders and print the amount of files in those folders in that order
	for _, folder := range folderSlice {

		filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
			if info.Mode().IsRegular() {
				booksMap[info.Size()] = append(booksMap[info.Size()], path)
			}

			return err
		})
	}
}

func removeUniqueEntries(booksMap map[int64][]string) {
	//remove from map entries where file size is unique
	for i, v := range booksMap {
		if len(v) == 1 {
			delete(booksMap, i)
		}
	}
}

func calculateFileChecksum(filepath string) chan []byte {

	hashCh := make(chan []byte)

	go func() {
		openFileLimitChan <- true

		file, err := os.Open(filepath)
		if err != nil {
			panic(err)
		}
		defer file.Close()

		const bufferSize = 10 * 1024 * 1024
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
		hashCh <- hash.Sum(nil)
		close(hashCh)
		<- openFileLimitChan
	}()

	return hashCh
}


func handleFilesWithTheSameSize(filelist []string, wg *sync.WaitGroup) {
	defer wg.Done()

	filesCount := len(filelist)

	chanSlice := make([]chan []byte, filesCount)

	for i, file := range filelist {
		chanSlice[i] = calculateFileChecksum(file)
	}

	hashMap := make(map[string]int)

	for i, ch := range chanSlice {

		tmpString := string(<-ch)

		if _, ok := hashMap[tmpString]; !ok { //check if hash exists in map
			hashMap[tmpString] = 1
		} else {
			fmt.Printf(START_RED+" Removing %s \n"+END_COLOR, filelist[i])
			os.Remove(filelist[i])
		}
	}

}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	//folders to check in priority order
	folders := []string{"_Przeczytane", "Sorted", "Unsorted"}

	booksMap := make(map[int64][]string)

	populateMap(booksMap, folders)

	removeUniqueEntries(booksMap)

	for _, listOfFiles := range booksMap {
		wg.Add(1)

		go handleFilesWithTheSameSize(listOfFiles, &wg)
	}

	wg.Wait()
}
