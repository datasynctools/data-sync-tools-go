//Package syncutil provides reusable common utilities.
package syncutil

import (
	"fmt"
	"log"
	"runtime"
	"strings"
)

//GetCallingName obtains the name of the calling function
func GetCallingName() string {
	pc, _, _, _ := runtime.Caller(1)

	fullPCName := runtime.FuncForPC(pc).Name()
	lastIndexOfPc := strings.LastIndex(fullPCName, "/") + 1
	justPcName := fullPCName[lastIndexOfPc:len(fullPCName)]
	lastIndexOfJustName := strings.LastIndex(justPcName, ".") + 1
	justName := justPcName[lastIndexOfJustName:len(justPcName)]

	return justName
}

//Info prints an info msg to log.
func Info(msg ...interface{}) {
	pc, file, line, _ := runtime.Caller(1)

	fullPCName := runtime.FuncForPC(pc).Name()
	lastIndexOfPc := strings.LastIndex(fullPCName, "/") + 1
	justPcName := fullPCName[lastIndexOfPc:len(fullPCName)]

	lastIndexOfFile := strings.LastIndex(file, "/") + 1
	justFileName := file[lastIndexOfFile:len(file)]

	log.Printf("INFO [%s:%d] [%s] %v", justFileName, line, justPcName, msg)
}

//Error prints an error msg to log.
func Error(msg ...interface{}) {
	pc, file, line, _ := runtime.Caller(1)

	fullPCName := runtime.FuncForPC(pc).Name()
	lastIndexOfPc := strings.LastIndex(fullPCName, "/") + 1
	justPcName := fullPCName[lastIndexOfPc:len(fullPCName)]

	lastIndexOfFile := strings.LastIndex(file, "/") + 1
	justFileName := file[lastIndexOfFile:len(file)]

	log.Printf("ERROR [%s:%d] [%s] %v", justFileName, line, justPcName, msg)
}

//Warn prints an warn msg to log.
func Warn(msg ...interface{}) {
	pc, file, line, _ := runtime.Caller(1)

	fullPCName := runtime.FuncForPC(pc).Name()
	lastIndexOfPc := strings.LastIndex(fullPCName, "/") + 1
	justPcName := fullPCName[lastIndexOfPc:len(fullPCName)]

	lastIndexOfFile := strings.LastIndex(file, "/") + 1
	justFileName := file[lastIndexOfFile:len(file)]

	log.Printf("WARN [%s:%d] [%s] %v", justFileName, line, justPcName, msg)
}

//Debug prints an debug msg to log.
func Debug(msg ...interface{}) {
	pc, file, line, _ := runtime.Caller(1)

	fullPCName := runtime.FuncForPC(pc).Name()
	lastIndexOfPc := strings.LastIndex(fullPCName, "/") + 1
	justPcName := fullPCName[lastIndexOfPc:len(fullPCName)]

	lastIndexOfFile := strings.LastIndex(file, "/") + 1
	justFileName := file[lastIndexOfFile:len(file)]

	log.Printf("DEBUG [%s:%d] [%s] %v", justFileName, line, justPcName, msg)
}

//Fatal prints a fatal msg to log then panics.
func Fatal(msg ...interface{}) {
	pc, file, line, _ := runtime.Caller(1)

	fullPCName := runtime.FuncForPC(pc).Name()
	lastIndexOfPc := strings.LastIndex(fullPCName, "/") + 1
	justPcName := fullPCName[lastIndexOfPc:len(fullPCName)]

	lastIndexOfFile := strings.LastIndex(file, "/") + 1
	justFileName := file[lastIndexOfFile:len(file)]

	log.Printf("FATAL [%s:%d] [%s] %v", justFileName, line, justPcName, msg)
	panic("Fatal Error: " + fmt.Sprintf("FATAL [%s:%d] [%s] %v", justFileName, line, justPcName, msg))
}

//NotImplementedMsg prints an NotImplemented msg to log.
func NotImplementedMsg(msg ...interface{}) {
	pc, file, line, _ := runtime.Caller(1)

	fullPCName := runtime.FuncForPC(pc).Name()
	lastIndexOfPc := strings.LastIndex(fullPCName, "/") + 1
	justPcName := fullPCName[lastIndexOfPc:len(fullPCName)]

	lastIndexOfFile := strings.LastIndex(file, "/") + 1
	justFileName := file[lastIndexOfFile:len(file)]

	log.Printf("MISSING [%s:%d] [%s] %v", justFileName, line, justPcName, msg)
}
