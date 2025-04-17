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

func Zip(path string, profile string, values map[string]interface{}) {
	logger := logging.GetLogManager()
	today := time.Now().Format("20060102")
	profile = filepath.Clean(strings.ReplaceAll(profile, string(filepath.Separator), "-"))
	filePtr, err := os.Create(
		filepath.Join(
			filepath.Clean(path),
			fmt.Sprintf("nuvola-%s_%s.zip", profile, today)),
	)
	if err != nil {
		logger.Error("Error on creating output folder", "err", err)
	}
	defer func() {
		if cerr := filePtr.Close(); cerr != nil {
			logger.Error("Error closing file", "err", err)
		}
	}()

	zipWriter := zip.NewWriter(filePtr)
	defer func() {
		if cerr := zipWriter.Close(); cerr != nil {
			logger.Error("Error closing zip writer", "err", cerr)
		}
	}()

	for key, value := range values {
		writer, err := zipWriter.Create(fmt.Sprintf("%s_%s.json", key, today))
		if err != nil {
			logger.Error("Error on creating file", "err", err)
		}

		data := logger.PrettyJSON(value)
		if _, err := writer.Write(data); err != nil {
			logger.Error("Error writing file content", "err", err)
		}
	}
}

func UnzipInMemory(zipfile string) *zip.ReadCloser {
	r, err := zip.OpenReader(zipfile)
	if err != nil {
		logging.GetLogManager().Error("Error on opening ZIP file", "err", err)
	}
	return r
}
