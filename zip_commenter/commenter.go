package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"io"
	"os"
	"time"
)

const (
	CommentBytesLimit = 65535
)

var (
	commentFile string
	inputFile   string
	outputFile  string
	mode        string
)

func init() {
	flag.StringVar(&inputFile, "i", "", "path to the input zip file, not null")
	flag.StringVar(&commentFile, "c", "", "path to the comment file, should not be larger than 65535 bytes")
	flag.StringVar(&outputFile, "o", "./output.zip", "path to the output file")
	flag.StringVar(&mode, "m", "w", "mode, 'w' will write comment to zip, 'r' will read and print out comment")
}

func writeComment() {
	commentBytes, _ := os.ReadFile(commentFile)
	if len(commentBytes) > CommentBytesLimit {
		flag.Usage()
		return
	}
	fmt.Printf("comment has %d bytes\n", len(commentBytes))

	reader, err := zip.OpenReader(inputFile)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer reader.Close()

	fmt.Printf("got %d files\n", len(reader.File))

	// Create a new zip writer
	outputF, err := os.Create(outputFile)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer outputF.Close()

	zipWriter := zip.NewWriter(outputF)

	// Copy each file from the existing zip to the new zip writer
	for _, file := range reader.File {
		// Open the file from the existing zip
		inputFile, err := file.Open()
		if err != nil {
			fmt.Println(err)
			return
		}
		defer inputFile.Close()

		// Create a new file entry in the new zip
		writer, err := zipWriter.CreateHeader(&file.FileHeader)
		if err != nil {
			fmt.Println(err)
			return
		}

		// Copy the file content to the new zip
		_, err = io.Copy(writer, inputFile)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	// Add information to the zip file header
	zipWriter.SetComment(string(commentBytes))

	// Close the zip writer
	err = zipWriter.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
}

func readComment() {
	reader, err := zip.OpenReader(inputFile)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer reader.Close()

	if reader.Comment == "" {
		fmt.Println("No comment!")
	} else {
		fmt.Printf("Comment: \n%s\n", reader.Comment)
	}
}

func main() {
	start := time.Now()

	flag.Usage = func() {
		fmt.Println(`ZipCommenter, a tool to write/read comment info to/of an existing zip file and output a new zip, e.g.
 Write comment to zip:
	zip_commenter -i foo.zip -c meta.json -m w
 Read comment of zip:
	zip_commenter -i foo.zip -m r`)
		flag.PrintDefaults()
	}

	flag.Parse()

	if inputFile == "" || (mode == "w" && commentFile == "") {
		flag.Usage()
		return
	}

	switch mode {
	case "w":
		writeComment()
	case "r":
		readComment()
	}

	fmt.Printf("processed successfully, took %s\n", time.Since(start).String())
}
