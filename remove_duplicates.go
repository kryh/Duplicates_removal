package main

import "fmt"
import "os"
import "path/filepath"
import "io/ioutil"
import "crypto/md5"

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

func calculateChecksum(filename string) string {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println("Error during opening " + filename)
		return ""
	}
	h := md5.New()
	h.Write(data)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func main() {
 	addBooksToMap("_Przeczytane")
	addBooksToMap("_Przeczytane_do_przejrzenia")
	addBooksToMap("Sorted")
	addBooksToMap("Unsorted")
	
	mapHashFilename := make(map[string]string)
	
	for _, names := range booksMap {
		if len(names) > 1 {
			for /*index*/_, bookname := range names {
				calculatedHash := calculateChecksum(bookname)
				if _,ok := mapHashFilename[calculatedHash]; !ok {  //check if hash already exists
					mapHashFilename[calculatedHash] = bookname
				} else {
					fmt.Println("Removing book: ", bookname)
					os.Remove(bookname)
				}
			}
		}
	}
}
