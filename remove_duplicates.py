#!/usr/bin/python
import os
import hashlib

if __name__ == '__main__':
	
	def calculateHashValue(fileName):
		fileHandle = open(fileName)
		fileContent = fileHandle.read()
		fileHandle.close()
		calculatedHash = hashlib.md5(fileContent).hexdigest()
		return calculatedHash

	def addBookToDictionary(dictionary, rootdir):
		for directory, subFolders, books in os.walk(rootdir):
			for file in books:
				book = os.path.join(directory,file)
				fileSize = os.path.getsize(book)
				if fileSize in dictionary.keys():
					dictionary[fileSize].append(book)
				else:
					dictionary[fileSize] = [book]
		
		
	#dictionary wielkosc pliku - lista nazw plikow o tej wielkosci
	allBooks = {}
	addBookToDictionary(allBooks, "./_Przeczytane")
	addBookToDictionary(allBooks, "./_Przeczytane_do_przejrzenia")
	addBookToDictionary(allBooks, "./Sorted")
	addBookToDictionary(allBooks, "./Unsorted")


	dict_hash_filename = {}
	savedFileSize = 0
	for fileSize, listOfFilenames in allBooks.items():
		if len(listOfFilenames)>1:		
			for fileName in listOfFilenames:
				calculatedHash = calculateHashValue(fileName)

				print calculatedHash, fileSize, fileName
				
				if "_Przeczytane" in fileName: #this file cannot be deleted
					dict_hash_filename[calculatedHash] = fileName
				elif calculatedHash not in dict_hash_filename.keys():
					dict_hash_filename[calculatedHash] = fileName
				else:
					print "  Deleting file: %s\n" %(fileName)
					os.unlink(fileName)
					savedFileSize += fileSize
				
	print "\n  Deleted %d kB\n" %(savedFileSize/1024)
