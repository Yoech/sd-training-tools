package main

import (
	"Yoech.com/Modules/CCCommon"
	"flag"
	_ "go.uber.org/automaxprocs"
	"log"
	"runtime"
)

var (
	inPath  string
	outPath string
	fileExt = ""
	prefix  = ""
	split   = ""
	split2  = ""
)

func init() {
	flag.StringVar(&inPath, "i", "./input", "specify the input folder")
	flag.StringVar(&outPath, "o", "", "specify the output folder,using same input folder if [-o] ignored")
	flag.StringVar(&fileExt, "e", "txt", "specify the file extension,support [txt|png]")
	flag.StringVar(&prefix, "p", "Prefix_", "specify the content prefix you want appended")
	flag.StringVar(&split, "s", "-", "specify the file name spliter")
	flag.StringVar(&split2, "s2", "_", "specify the file name spliter second")
}

func main() {
	// using uber automaxprocs for replace update GOMAXPROCS
	log.Printf("uber automaxprocs => GOMAXPROCS[%v]", runtime.GOMAXPROCS(-1))
	// runtime.GOMAXPROCS(runtime.NumCPU())

	flag.Parse()

	CCCommon.LogEnabled = false
	defer CCCommon.PanicHandler()

	CCCommon.Logger.Info("try to start stable-diffusion-webui-tools ...")

	if !verify() {
		flag.Usage()
		return
	}

	st := CCCommon.TimeMillSecond()
	if err := ProcessFolder(inPath); err != nil {
		CCCommon.Logger.Errorf("Exception traversing input folder!\nerror is:\n%v", err)
		return
	}
	cost := CCCommon.TimeMillSecond() - st
	CCCommon.LogEnabled = true
	CCCommon.Logger.Infof("Cost[%v]ms ... OK", cost)
}

func verify() bool {
	if inPath == "" {
		CCCommon.Logger.Errorf("Please specify the input folder!")
		return false
	}

	if outPath == "" {
		outPath = inPath
	}

	if fileExt == "" {
		CCCommon.Logger.Errorf("Please specify the file ext!")
		return false
	}

	if prefix == "" {
		CCCommon.Logger.Errorf("Please specify the content prefix you want appended!")
		return false
	}

	CCCommon.Logger.Infof("verify ... OK")
	return true
}
