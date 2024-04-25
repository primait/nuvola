package zip

import (
	"archive/zip"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/primait/nuvola/pkg/io/logging"
)

func Zip(path string, profile string, values *map[string]interface{}) {
	today := time.Now().Format("20060102")
	fileSeparator := string(filepath.Separator)
	profile = filepath.Clean(strings.Replace(profile, fileSeparator, "-", -1))
	filePtr, err := os.Create(fmt.Sprintf("%s%snuvola-%s_%s.zip", filepath.Clean(path), fileSeparator, profile, today))
	if err != nil {
		logging.HandleError(err, "Zip", "Error on creating output folder")
	}
	defer func() {
		if err := filePtr.Close(); err != nil {
			logging.HandleError(err, "Zip", "Error closing file")
		}
	}()

	MyZipWriter := zip.NewWriter(filePtr)
	defer MyZipWriter.Close()

	for key, value := range *values {
		writer, err := MyZipWriter.Create(fmt.Sprintf("%s_%s.json", key, today))
		if err != nil {
			fmt.Println(err)
		}

		_, err = writer.Write(logging.PrettyJSON(value))
		if err != nil {
			logging.HandleError(err, "Zip", "Error on writing file content")
		}
	}
}

func UnzipInMemory(zipfile string) (r *zip.ReadCloser) {
	r, err := zip.OpenReader(zipfile)
	if err != nil {
		logging.HandleError(err, "Zip", "Error on opening ZIP file")
	}
	return
}
