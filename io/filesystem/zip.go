package filesystem

import (
	"archive/zip"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	print "nuvola/io/cli/output"
)

func ZipIt(path string, profile string, values *map[string]interface{}) {
	today := time.Now().Format("20060102")
	filePtr, err := os.Create(fmt.Sprintf("%s%snuvola-%s_%s.zip", path, string(filepath.Separator), profile, today))
	if err != nil {
		fmt.Println(err)
	}
	defer func() {
		if err := filePtr.Close(); err != nil {
			log.Fatalf("Error closing file: %s\n", err)
		}
	}()

	MyZipWriter := zip.NewWriter(filePtr)
	defer MyZipWriter.Close()

	for key, value := range *values {
		writer, err := MyZipWriter.Create(fmt.Sprintf("%s_%s.json", key, today))
		if err != nil {
			fmt.Println(err)
		}

		_, err = writer.Write([]byte(print.PrettyJson(value)))
		if err != nil {
			fmt.Println(err)
		}
	}
}

func UnzipInMemory(zipfile string) (r *zip.ReadCloser) {
	r, err := zip.OpenReader(zipfile)
	if err != nil {
		log.Fatal(err)
	}
	return
}
