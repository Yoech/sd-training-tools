package CCCommon

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	"runtime/pprof"

	"io"
	"strings"
	"time"

	"github.com/op/go-logging"
	"github.com/robfig/cron"
)

var (
	logFile    *os.File
	LogEnabled = false
)

func WriteLog(logFile string) *os.File {
	logfile, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		os.Exit(-1)
	}

	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Llongfile)
	log.SetOutput(logfile)
	return logfile
}

func WaitInput() {
	var cmd string
	for {
		_, err := fmt.Scanf("%s\n", &cmd)
		if err != nil {
			// log.Println("Scanf err:", err)
			continue
		}

		switch cmd {
		case "exit", "quit":
			log.Println("exit by user")
			return
		case "gr":
			log.Println("current goroutine count:", runtime.NumGoroutine())
			break
		case "gd":
			_ = pprof.Lookup("goroutine").WriteTo(os.Stdout, 1)
			break
		default:
			break
		}
	}
}

func WaitForSignal() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	// Block until a signal is received.
	s := <-c
	fmt.Println("Got signal:", s)
}

func StartLogFileOutPut() {
	writer := createWriters()
	log.SetOutput(writer)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	task := cron.New()
	spec := "0 0 0 * * *"
	_ = task.AddFunc(spec, func() {
		writer = createWriters()
		log.SetOutput(writer)
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	})
	task.Start()
}

func createWriters() io.Writer {
	if logFile != nil {
		_ = logFile.Close()
		logFile = nil
	}
	path := os.Args[0]
	startIndex := strings.LastIndex(path, "/")
	if startIndex < 0 {
		startIndex = strings.LastIndex(path, "\\")
		if startIndex < 0 {
			startIndex = 0
		}
	}
	fileName := path[startIndex:strings.LastIndex(path, ".")]
	dirName := path[:startIndex]
	dirName += "log/" + fileName
	err := os.MkdirAll(dirName, os.ModePerm)
	if err != nil {
		return os.Stdout
	}

	year, month, day := time.Now().Date()
	logName := dirName + fmt.Sprintf("/%04d-%02d-%02d.log", year, month, day)
	logFile, err = os.OpenFile(logName, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		return os.Stdout
	}
	os.Stderr = logFile

	writers := []io.Writer{
		logFile,
		os.Stdout,
	}

	return io.MultiWriter(writers...)
}

func createLoggingWriters() []io.Writer {
	if logFile != nil {
		_ = logFile.Close()
		logFile = nil
	}
	path := os.Args[0]

	startIndex := strings.LastIndex(path, "/")
	if startIndex < 0 {
		startIndex = strings.LastIndex(path, "\\")
		if startIndex < 0 {
			startIndex = 0
		}
	}

	fileName := path[startIndex+1:]

	dirName := path[:startIndex]
	dirName += "/log/" + fileName
	err := os.MkdirAll(dirName, os.ModePerm)
	if err != nil {
		return []io.Writer{os.Stdout}
	}

	year, month, day := time.Now().Date()
	logName := dirName + fmt.Sprintf("/%04d-%02d-%02d.log", year, month, day)
	logFile, err = os.OpenFile(logName, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		return []io.Writer{os.Stdout}
	}
	os.Stderr = logFile

	writers := []io.Writer{
		logFile,
		os.Stdout,
	}

	return writers
}

type LogTag struct {
	IsInit    bool
	LogHandle *logging.Logger
}

var Logger LogTag

func LogInit() {
	if !LogEnabled {
		return
	}
	Logger.LogHandle = logging.MustGetLogger("CCServer")
	format := logging.MustStringFormatter(
		`%{time:2006-01-02 15:04:05.999} [%{level}] <%{longfile}> : [%{callpath:1}] %{message}`,
	)
	formatstd := logging.MustStringFormatter(
		`%{time:2006-01-02 15:04:05.999} [%{level}] <%{shortfile}> : %{message}`,
	)

	Logger.LogHandle.ExtraCalldepth = 1
	writer := createLoggingWriters()

	backendlog := logging.NewLogBackend(writer[0], "", 0)
	backendstd := logging.NewLogBackend(writer[1], "", 0)
	backend1Formatter := logging.NewBackendFormatter(backendlog, format)
	backend2Formatter := logging.NewBackendFormatter(backendstd, formatstd)
	be := logging.SetBackend(backend1Formatter, backend2Formatter)
	Logger.LogHandle.SetBackend(be)

	task := cron.New()
	spec := "0 0 0 * * *"
	_ = task.AddFunc(spec, func() {
		writer = createLoggingWriters()
		backendlog = logging.NewLogBackend(writer[0], "", 0)
		backendstd = logging.NewLogBackend(writer[1], "", 0)
		backend1Formatter = logging.NewBackendFormatter(backendlog, format)
		backend2Formatter = logging.NewBackendFormatter(backendstd, formatstd)
		be = logging.SetBackend(backend1Formatter, backend2Formatter)
		Logger.LogHandle.SetBackend(be)
	})
	task.Start()

	Logger.IsInit = true
}

func (l *LogTag) Panic(args ...interface{}) {
	if !l.IsInit {
		log.Panic(args...)
	} else {
		l.LogHandle.Panic(args...)
	}

	// pc, file, line, _ := runtime.Caller(1)
	// functag := runtime.FuncForPC(pc)
	// content := fmt.Sprint(args...)
	// theme := fmt.Sprintf("%s:%d [%s]", file, line, functag.Name())
	//
	// go func(t, c string) {
	//	SendMail(param.MailTo, theme, content)
	// }(theme, content)
}

func (l *LogTag) Panicf(format string, args ...interface{}) {
	if !l.IsInit {
		log.Panicf(format, args...)
	} else {
		l.LogHandle.Panicf(format, args...)
	}

	// pc, file, line, _ := runtime.Caller(1)
	// functag := runtime.FuncForPC(pc)
	// content := fmt.Sprintf(format, args...)
	// theme := fmt.Sprintf("%s:%d [%s]", file, line, functag.Name())
	//
	// go func(t, c string) {
	//	SendMail(param.MailTo, theme, content)
	// }(theme, content)
}

// Critical
func (l *LogTag) Critical(args ...interface{}) {
	if !l.IsInit {
		log.Println(args...)
	} else {
		l.LogHandle.Critical(args...)
	}

	// pc, file, line, _ := runtime.Caller(1)
	// functag := runtime.FuncForPC(pc)
	// content := fmt.Sprint(args...)
	// theme := fmt.Sprintf("%s:%d [%s]", file, line, functag.Name())
	//
	// go func(t, c string) {
	//	SendMail(param.MailTo, theme, content)
	// }(theme, content)
}

func (l *LogTag) Criticalf(format string, args ...interface{}) {
	if !l.IsInit {
		log.Printf(format, args...)
	} else {
		l.LogHandle.Criticalf(format, args...)
	}

	// pc, file, line, _ := runtime.Caller(1)
	// functag := runtime.FuncForPC(pc)
	// content := fmt.Sprintf(format, args...)
	// theme := fmt.Sprintf("%s:%d [%s]", file, line, functag.Name())
	//
	// go func(t, c string) {
	//	SendMail(param.MailTo, theme, content)
	// }(theme, content)
}

func (l *LogTag) Error(args ...interface{}) {
	if !LogEnabled {
		return
	}
	if !l.IsInit {
		log.Println(args...)
	} else {
		l.LogHandle.Error(args...)
	}
}

func (l *LogTag) Errorf(format string, args ...interface{}) {
	if !LogEnabled {
		return
	}
	if !l.IsInit {
		log.Printf(format, args...)
	} else {
		l.LogHandle.Errorf(format, args...)
	}
}

func (l *LogTag) Info(args ...interface{}) {
	if !LogEnabled {
		return
	}
	if !l.IsInit {
		log.Println(args...)
	} else {
		l.LogHandle.Info(args...)
	}
}

func (l *LogTag) Infof(format string, args ...interface{}) {
	if !LogEnabled {
		return
	}
	if !l.IsInit {
		log.Printf(format, args...)
	} else {
		l.LogHandle.Infof(format, args...)
	}
}

func (l *LogTag) Debug(args ...interface{}) {
	if !LogEnabled {
		return
	}
	if !l.IsInit {
		log.Println(args...)
	} else {
		l.LogHandle.Debug(args...)
	}
}

func (l *LogTag) Debugf(format string, args ...interface{}) {
	if !LogEnabled {
		return
	}
	if !l.IsInit {
		log.Printf(format, args...)
	} else {
		l.LogHandle.Debugf(format, args...)
	}
}

func PanicHandler() {
	var errinfo string
	var stackinfo string

	enter := false
	if err := recover(); err != nil {
		enter = true
		errinfo = fmt.Sprintf("%v\r\n", err)
		Logger.Error(errinfo)
	}

	if !enter {
		return
	}

	stackinfo = string(debug.Stack())
	Logger.Error(stackinfo)

	// calldepth = 4  暂时先这样吧
	pc, file, line, _ := runtime.Caller(4)
	functag := runtime.FuncForPC(pc)
	content := fmt.Sprintf("%s\r\n%s", errinfo, stackinfo)
	theme := fmt.Sprintf("%s:%d [%s]", file, line, functag.Name())

	Logger.Errorf("PanicHandler.functag[%v].theme[%v].content[%v]", functag, theme, content)

	// // 这里一定要用defer
	// defer func(t, c string) {
	//	SendMail(param.MailTo, theme, content)
	// }(theme, content)
}
