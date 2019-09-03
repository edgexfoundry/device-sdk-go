package mock

import (
	"io/ioutil"
	"path/filepath"
)

func loadData(path string) (map[string][]byte, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	fileInfo, err := ioutil.ReadDir(absPath)
	if err != nil {
		return nil, err
	}

	result := make(map[string][]byte, len(fileInfo))
	for _, file := range fileInfo {
		fName := file.Name()
		ext := filepath.Ext(fName)
		name := fName[0 : len(fName)-len(ext)]
		fullPath := absPath + "/" + fName
		file, err := ioutil.ReadFile(fullPath)
		if err != nil {
			return nil, err
		}
		result[name] = file
	}

	return result, nil
}
