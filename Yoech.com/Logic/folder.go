package main

import (
	"Yoech.com/Modules/CCCommon"
	"fmt"
	"os"
	"strings"
)

func ProcessFolder(pathname string) error {
	rd, err := os.ReadDir(pathname)
	for _, fi := range rd {
		if fi.IsDir() {
			fmt.Printf("[%s]\n", pathname+"//"+fi.Name())
			_ = ProcessFolder(pathname + "//" + fi.Name())
		} else {
			ProcessFile(pathname, fi.Name())
		}
	}
	return err
}

func ProcessFile(folder string, filePath string) bool {
	var fullPath = folder + "//" + filePath
	var arr []string
	if arr = strings.Split(filePath, "."); len(arr) != 2 {
		return false
	}

	// 00039-1-Gambit_Foreground.txt
	if strings.Compare(strings.ToLower(arr[1]), strings.ToLower(fileExt)) != 0 {
		return false
	}

	// 00039-1-Gambit_Foreground
	arr2 := strings.Split(arr[0], split)
	if len(arr2) != 3 {
		CCCommon.Logger.Infof("ProcessFile.fullPath[%v].split[%v].invalid", fullPath, split)
		return false
	}

	// Gambit_Foreground
	arr3 := strings.Split(arr2[2], split2)
	if len(arr3) != 2 {
		CCCommon.Logger.Infof("ProcessFile.fullPath[%v].split2[%v].invalid", fullPath, split2)
		return false
	}

	if strings.ToLower(fileExt) == "png" {
		fullPath = folder + "//" + arr[0] + ".txt"
	}

	_ = os.Remove(fullPath)

	var err error
	var f *os.File
	if f, err = os.Create(fullPath); err != nil {
		CCCommon.Logger.Errorf("ProcessFile.Create[%v].err[%v]", fullPath, err)
		return false
	}
	// Prefix_Gambit
	_, _ = f.WriteString(prefix + arr3[0])
	_ = f.Close()
	CCCommon.Logger.Infof("ProcessFile.fullPath[%v] ... OK", fullPath)

	return true
}
