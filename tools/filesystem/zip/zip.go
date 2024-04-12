package zip

import (
	"archive/zip"
	"fmt"
	"os"
	"path/filepath"
	"time"

	print "github.com/primait/nuvola/tools/cli/output"
	nuvolaerror "github.com/primait/nuvola/tools/error"
)

func Zip(path string, profile string, values *map[string]interface{}) {
	today := time.Now().Format("20060102")
	filePtr, err := os.Create(fmt.Sprintf("%s%snuvola-%s_%s.zip", path, string(filepath.Separator), profile, today))
	if err != nil {
		nuvolaerror.HandleError(err, "Zip", "Error on creating output folder")
	}
	defer func() {
		if err := filePtr.Close(); err != nil {
			nuvolaerror.HandleError(err, "Zip", "Error closing file")
		}
	}()

	MyZipWriter := zip.NewWriter(filePtr)
	defer MyZipWriter.Close()

	for key, value := range *values {
		writer, err := MyZipWriter.Create(fmt.Sprintf("%s_%s.json", key, today))
		if err != nil {
			fmt.Println(err)
		}

		_, err = writer.Write(print.PrettyJSON(value))
		if err != nil {
			nuvolaerror.HandleError(err, "Zip", "Error on writing file content")
		}
	}
}

func UnzipInMemory(zipfile string) (r *zip.ReadCloser) {
	r, err := zip.OpenReader(zipfile)
	if err != nil {
		nuvolaerror.HandleError(err, "Zip", "Error on opening ZIP file")
	}
	return
}
