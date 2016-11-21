package cclog

import (
	"time"
	"runtime"
	"path"
	"strings"
	"fmt"
	"os"
	"errors"
)



// Debug, Debug, Info, Warn, Error, Critical
const (
	levelDebug = iota + 1   // Debug message
	levelInfo               // Info message
	levelWarn               // Warn message
	levelError              // Error message
	levelFatal              // Fatal message(not)
)



var logChan = make(chan string, 10000)

var enableFile = false
var enableConsole = true
var zipLogFile = false

var logLevel = levelDebug


var logFormater = "%Time [%Level] %Msg --[%Line]%File"

var logfile *os.File


// safe for read
var levelString = make(map[int]string)
var stringLevel = make(map[string]int)


func init() {
	levelString[levelDebug] = "DEBU"
	levelString[levelInfo] = "INFO"
	levelString[levelWarn] = "WARN"
	levelString[levelError] = "ERRO"
	levelString[levelFatal] = "FATA"

	stringLevel["DEBU"] = levelDebug
	stringLevel["INFO"] = levelInfo
	stringLevel["WARN"] = levelWarn
	stringLevel["ERRO"] = levelError
	stringLevel["FATA"] = levelFatal

	go rotate()

	go writeLogFile()
}





func SetLevel(level string) error {

	if l,ok := stringLevel[level];ok {
		logLevel = l
		return nil
	}
	return errors.New("unsupport level:"+level)
}



func EnableFile() {
	enableFile = true
}

func ZipLog() {
	zipLogFile = true
}

func DisableConsole() {
	enableConsole = false
}

func Debug(msg ...interface{})  {
	write(levelDebug,msg)
}

func Debugf(format string, a ...interface{})  {
	msg := []interface{} {fmt.Sprintf(format, a...)}
	write(levelDebug,msg)
}

func Info(msg ...interface{})  {
	write(levelInfo,msg)
}

func Infof(format string, a ...interface{})  {
	msg := []interface{} {fmt.Sprintf(format, a...)}
	write(levelInfo,msg)
}

func Warn(msg ...interface{})  {
	write(levelWarn,msg)
}

func Warnf(format string, a ...interface{})  {
	msg := []interface{} {fmt.Sprintf(format, a...)}
	write(levelWarn,msg)
}

func Error(msg ...interface{})  {
	write(levelError,msg)
}

func Errorf(format string, a ...interface{})  {
	msg := []interface{} {fmt.Sprintf(format, a...)}
	write(levelError,msg)
}

func Fatal(msg ...interface{})  {
	write(levelFatal,msg)
}

func Fatalf(format string, a ...interface{})  {
	msg := []interface{} {fmt.Sprintf(format, a...)}
	write(levelFatal,msg)
}




// generate the log struct and push to logChan, then wait to be write
func write(level int,msg... interface{})  {
	if level<logLevel {
		return
	}

	_,file,line,ok := runtime.Caller(2)
	if !ok {
		file = "???"
		line = 0
	}
	_,filename := path.Split(file)

	fmt_msg := logFormater
	fmt_msg = strings.Replace(fmt_msg, "%Time",time.Now().Format("20060102-15:04:05"),-1)
	fmt_msg = strings.Replace(fmt_msg, "%Level",levelString[level],-1)
	fmt_msg = strings.Replace(fmt_msg, "%Line",fmt.Sprint(line),-1)
	fmt_msg = strings.Replace(fmt_msg, "%File",filename,-1)

	msgs := msg[0].([]interface{})
	msg_str := ""
	for k,_ := range msgs {
		msg_str = msg_str +fmt.Sprintf("%+v ",msgs[k])
	}
	fmt_msg = strings.Replace(fmt_msg, "%Msg",msg_str,-1) + "\n"

	// 控制console打印
	if enableConsole{
		consoleLog(level, fmt_msg)
	}

	logChan <- fmt_msg
}

// ************color console log*************************************
var color = []string{
	"1;0",		// none		none
	"1;30",		// debug	gray
	"1;34",		// info		blue
	"1;36",		// warn
	"1;31",		// error	red
	"1;35",		// fatal	magenta
}

func consoleLog(level int, content string) {
	os.Stdout.Write([]byte("\033["+color[level]+"m"+content+"\033[0m"))
}
// ******************************************************************



// ************file log**********************************************
// log文件按天切割
func rotate()  {

	now := time.Now()
	day := time.Date(now.Year(),now.Month(),now.Day(),0,0,0,0,time.Local)

	err := os.MkdirAll("log", 0766)
	if err != nil {
		panic("创建目录log失败:"+err.Error())
	}

	for {
		// 每天两点切换日志，存放在当前目录的log目录下
		tmp,err := os.OpenFile("./log/"+day.Format("20060102")+".log", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0766)
		if err != nil {
			fmt.Println("打开或创建文件失败:"+err.Error())
			panic("打开或创建文件失败:"+err.Error())
		}
		logfile = tmp

		day = day.Add(24*time.Hour)

		time.Sleep(day.Sub(time.Now()))
	}

}

func writeLogFile()  {
	for{
		select {
		case log := <-logChan:
			if enableFile{
				_, err := logfile.WriteString(log)
				if err != nil {
					fmt.Printf("写日志文件出错:%T, %v\n",log, err)
				}
			}
		}
	}
	logfile.Close()
}

// ******************************************************************

