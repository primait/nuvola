package files

import (
	"io/fs"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/primait/nuvola/pkg/io/logging"
)

func PrettyJSONToFile(filePath string, fileName string, s interface{}) {
	logger := logging.GetLogManager()
	if err := os.MkdirAll(filePath, os.FileMode(0775)); err != nil {
		logger.Error("Error on creating/reading output folder", "err", err)
	}

	filePath = filePath + string(filepath.Separator) + fileName
	if err := os.WriteFile(filePath, logger.PrettyJSON(s), 0600); err != nil {
		logger.Error("Error on writing file", "err", err)
	}
}

func GetFiles(root, pattern string) []string {
	var a []string
	err := filepath.WalkDir(NormalizePath(root), func(s string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if ok, err := regexp.Match(pattern, []byte(filepath.Ext(d.Name()))); ok {
			a = append(a, s)
		} else {
			return err
		}
		return nil
	})
	if err != nil {
		logging.GetLogManager().Error("Error on reading file", "err", err)
	}
	return a
}

func NormalizePath(path string) string {
	usr, _ := user.Current()
	dir := usr.HomeDir
	if path == "~" {
		path = dir
	} else if strings.HasPrefix(path, "~/") {
		path = filepath.Join(dir, path[2:])
	}

	path, _ = filepath.Abs(filepath.Clean(path))
	return path
}
