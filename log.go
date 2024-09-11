package log

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/chzyer/readline"
	"github.com/fatih/color"
)

var stdout io.Writer = color.Output
var g_rl *readline.Instance = nil
var debug_output = true
var mtx_log *sync.Mutex = &sync.Mutex{}
var fileLogger *log.Logger
var currentLogFile *os.File
var lastLogDate time.Time
var logDir string = "logs"

const (
	DEBUG = iota
	INFO
	IMPORTANT
	WARNING
	ERROR
	FATAL
	SUCCESS
)

var LogLabels = map[int]string{
	DEBUG:     "dbg",
	INFO:      "inf",
	IMPORTANT: "imp",
	WARNING:   "war",
	ERROR:     "err",
	FATAL:     "!!!",
	SUCCESS:   "+++",
}

func init() {
	fmt.Println("Initializing logging system...")
	
	// Get the executable's directory
	exePath, err := os.Executable()
	if err != nil {
		fmt.Println("Error getting executable path:", err)
		return
	}
	exeDir := filepath.Dir(exePath)
	
	// Set the log directory relative to the executable
	logDir = filepath.Join(exeDir, "logs")
	
	err = os.MkdirAll(logDir, os.ModePerm)
	if err != nil {
		fmt.Println("Error creating logs directory:", err)
		return
	}
	
	err = createNewLogFile()
	if err != nil {
		fmt.Println("Error creating initial log file:", err)
		return
	}
	
	fmt.Printf("Logging system initialized. Logs will be saved in: %s\n", logDir)
}

func createNewLogFile() error {
	if currentLogFile != nil {
		currentLogFile.Close()
	}

	now := time.Now()
	filename := filepath.Join(logDir, fmt.Sprintf("log_%s.txt", now.Format("2006-01-02")))
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("error opening log file: %v", err)
	}

	currentLogFile = file
	fileLogger = log.New(file, "", log.Ltime)
	lastLogDate = now
	
	fmt.Printf("Created new log file: %s\n", filename)
	return nil
}

func checkRotateLog() {
	now := time.Now()
	if now.Day() != lastLogDate.Day() {
		err := createNewLogFile()
		if err != nil {
			fmt.Println("Error rotating log file:", err)
		}
	}
}

func DebugEnable(enable bool) {
	debug_output = enable
}

func SetOutput(o io.Writer) {
	stdout = o
}

func SetReadline(rl *readline.Instance) {
	g_rl = rl
}

func GetOutput() io.Writer {
	return stdout
}

func NullLogger() *log.Logger {
	return log.New(io.Discard, "", 0)
}

func refreshReadline() {
	if g_rl != nil {
		g_rl.Refresh()
	}
}

func writeToLog(level int, format string, args ...interface{}) {
	mtx_log.Lock()
	defer mtx_log.Unlock()

	checkRotateLog()

	msg := format_msg(level, format, args...)
	fmt.Fprint(stdout, msg)
	if fileLogger != nil {
		fileLogger.Print(stripAnsiColors(msg))
	} else {
		fmt.Println("Warning: fileLogger is nil, log not written to file")
	}
	refreshReadline()
}

func Debug(format string, args ...interface{}) {
	if debug_output {
		writeToLog(DEBUG, format+"\n", args...)
	}
}

func Info(format string, args ...interface{}) {
	writeToLog(INFO, format+"\n", args...)
}

func Important(format string, args ...interface{}) {
	writeToLog(IMPORTANT, format+"\n", args...)
}

func Warning(format string, args ...interface{}) {
	writeToLog(WARNING, format+"\n", args...)
}

func Error(format string, args ...interface{}) {
	writeToLog(ERROR, format+"\n", args...)
}

func Fatal(format string, args ...interface{}) {
	writeToLog(FATAL, format+"\n", args...)
}

func Success(format string, args ...interface{}) {
	writeToLog(SUCCESS, format+"\n", args...)
}

func Printf(format string, args ...interface{}) {
	mtx_log.Lock()
	defer mtx_log.Unlock()

	msg := fmt.Sprintf(format, args...)
	fmt.Fprint(stdout, msg)
	if fileLogger != nil {
		fileLogger.Print(stripAnsiColors(msg))
	} else {
		fmt.Println("Warning: fileLogger is nil, log not written to file")
	}
	refreshReadline()
}

func format_msg(lvl int, format string, args ...interface{}) string {
	t := time.Now()
	var sign, msg *color.Color
	switch lvl {
	case DEBUG:
		sign = color.New(color.FgBlack, color.BgHiBlack)
		msg = color.New(color.Reset, color.FgHiBlack)
	case INFO:
		sign = color.New(color.FgGreen, color.BgBlack)
		msg = color.New(color.Reset)
	case IMPORTANT:
		sign = color.New(color.FgWhite, color.BgHiBlue)
		msg = color.New(color.Reset)
	case WARNING:
		sign = color.New(color.FgHiYellow, color.BgBlack)
		msg = color.New(color.Reset)
	case ERROR:
		sign = color.New(color.FgWhite, color.BgRed)
		msg = color.New(color.Reset, color.FgRed)
	case FATAL:
		sign = color.New(color.FgBlack, color.BgRed)
		msg = color.New(color.Reset, color.FgRed, color.Bold)
	case SUCCESS:
		sign = color.New(color.FgWhite, color.BgGreen)
		msg = color.New(color.Reset, color.FgGreen)
	}
	time_clr := color.New(color.Reset)
	return "\r[" + time_clr.Sprintf("%02d:%02d:%02d", t.Hour(), t.Minute(), t.Second()) + "] [" + sign.Sprintf("%s", LogLabels[lvl]) + "] " + msg.Sprintf(format, args...)
}

func stripAnsiColors(s string) string {
	var result []rune
	inEscapeSeq := false
	for _, r := range s {
		if inEscapeSeq {
			if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
				inEscapeSeq = false
			}
		} else if r == '\x1b' {
			inEscapeSeq = true
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}
