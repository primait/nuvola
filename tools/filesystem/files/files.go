package files

import (
	"io/fs"
	"io/ioutil"
	cli "nuvola/tools/cli/output"
	nuvolaerror "nuvola/tools/error"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
)

func PrettyJSONToFile(filePath string, fileName string, s interface{}) {
	if err := os.MkdirAll(filePath, os.FileMode(0775)); err != nil {
		nuvolaerror.HandleError(err, "Files - PrettyJSONToFile", "Error on creating/reading output folder")
	}

	filePath = filePath + string(filepath.Separator) + fileName
	if err := ioutil.WriteFile(filePath, cli.PrettyJSON(s), 0600); err != nil {
		nuvolaerror.HandleError(err, "Files - PrettyJSONToFile", "Error on writing file")
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
		nuvolaerror.HandleError(err, "Files - GetFiles", "Error on reading file")
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
