package utils

import (
	"os"
	"strings"
)

func WorkInProjectPath(project string) {
	currentDir, _ := os.Getwd()
	filePathList := strings.Split(currentDir, "/")
	var index int
	for k, d := range filePathList {
		if d == project {
			index = k + 1
			break
		}
	}
	err := os.Chdir(strings.Join(filePathList[:index], "/"))
	if err != nil {
		panic(any(err))
	}
}
