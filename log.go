package cclog

import (
	"time"
	"io/ioutil"
	"encoding/json"
	"sync"
	"runtime"
	"path"
	"strings"
	"fmt"
	"os"
	"errors"
)

const DEFAULT_MSG_FORMAT = "%Time [%Level] %Msg --[%Line]%File"
const DEFAULT_FILE_FORMAT = "log/20060102.log"

// Trace, Debug, Info, Warn, Error, Critical
const (
	levelTrace = iota + 1   // Trace message
	levelDebug              // Debug message
	levelInfo               // Info message
	levelWarn               // Warn message
	levelError              // Error message
	levelCritical           // Critical message(not)
)


// log config
type config struct {

	// file output
	FileLevel string  `json:"fileLevel"`
	FileNameFormat string  `json:"fileNameFormater"`
	FileMsgFormat string  `json:"fileMsgFormater"`
	FileRotate bool  `json:"fileRotate"`
	FileRotateType string  `json:"fileRotateType"`   // rotate by date/size/line
	FileMaxRotate int64  `json:"fileMaxRotate"`   // max file to be saved
	FileMaxLine int  `json:"fileMaxLine"`  // max line to rotate file
	FileMaxSize int  `json:"fileMaxSize"`  // max size to rotate file


	// console output
	Console bool  `json:"console"`
	ConsoleLevel string  `json:"consoleLevel"`
	ConsoleMsgFormat string  `json:"consoleMsgFormat"`

	// email output
	Email string  `json:"email"`
	EmailLevel string  `json:"emailLevel"`
	EmailMsgFormat string  `json:"emailMsgFormat"`

	// remote log server
	Remote string  `json:"remote"`
	RemoteLevel string  `json:"remoteLevel"`
	RemoteMsgFormat string  `json:"remoteMsgFormat"`
}




// log full info
type log struct {
	Time time.Time
	Level int
	Msg []interface{}
	Line int
	FileName string
}

var logPool *sync.Pool



var lock sync.Mutex
var cfg config
var logChan = make(chan *log, 10000)
var signalChan = make(chan string)
var outputs = make(map[string]outPut)

var stop = false

var levelString = make(map[int]string)
var stringLevel = make(map[string]int)
var rotateType  = make(map[string]int)


func init() {
	levelString[levelTrace] = "Trace"
	levelString[levelDebug] = "Debug"
	levelString[levelInfo] = "Info"
	levelString[levelWarn] = "Warn"
	levelString[levelError] = "Error"
	levelString[levelCritical] = "Critical"

	stringLevel["Trace"] = levelTrace
	stringLevel["Debug"] = levelDebug
	stringLevel["Info"] = levelInfo
	stringLevel["Warn"] = levelWarn
	stringLevel["Error"] = levelError
	stringLevel["Critical"] = levelCritical

	rotateType["daily"] = 1
	rotateType["size"] = 2
	rotateType["line"] = 3


	// console log default
	cfg.Console = true
	cfg.ConsoleLevel = "Trace"
	cfg.ConsoleMsgFormat = DEFAULT_MSG_FORMAT

	// file log default
	cfg.FileNameFormat = DEFAULT_FILE_FORMAT
	cfg.FileLevel = "Trace"
	cfg.FileMsgFormat = DEFAULT_MSG_FORMAT
	cfg.FileRotateType = "daily"
	cfg.FileMaxRotate = 20
	cfg.FileRotate = true

	// email log default
	cfg.EmailLevel = "Trace"
	cfg.EmailMsgFormat = DEFAULT_MSG_FORMAT

	// remote log default
	cfg.RemoteLevel = "Trace"
	cfg.RemoteMsgFormat = DEFAULT_MSG_FORMAT


	// init the log pool
	logPool = &sync.Pool{
		New:func() interface{}{
			return &log{}
		},
	}

	// register all output
	registerLoggers()
	// start log writer server
	go startLogger()
}

func registerLoggers()  {
	lock.Lock()  // block all writer when update log writer

	for k,_ := range outputs{
		outputs[k].Close()
	}

	outputs["file"] = newFileLog()
	outputs["console"] = newConsoleLog()

	lock.Unlock()
}

func ConfigFromFile(name string) error {
	data,err := ioutil.ReadFile(name)
	if  err!=nil{
		return err
	}
	return ConfigFromByte(data)
}

func ConfigFromByte(data []byte) error {
	tmp := config{}
	if err := json.Unmarshal(data,&tmp);err !=nil{
		return err
	}

	// file log default
	if stringLevel[tmp.FileLevel] == 0 {
		fmt.Fprintln(os.Stdout,"unsupport level for file: "+tmp.FileLevel)
		tmp.FileLevel = "Warn"
	}
	if len(strings.TrimSpace(tmp.FileNameFormat)) == 0 {
		tmp.FileNameFormat = DEFAULT_FILE_FORMAT
	}
	if len(strings.TrimSpace(tmp.FileMsgFormat)) == 0 {
		tmp.FileMsgFormat = DEFAULT_MSG_FORMAT
	}
	if rotateType[tmp.FileRotateType] == 0 {
		tmp.FileRotateType = "daily"
	}
	if tmp.FileMaxRotate == 0{
		tmp.FileMaxRotate = 20
	}

	// console log default
	if stringLevel[tmp.ConsoleLevel] == 0{
		fmt.Fprintln(os.Stdout,"unsupport level for console: "+tmp.ConsoleLevel)
		tmp.ConsoleLevel = "Warn"
	}
	if len(strings.TrimSpace(tmp.ConsoleMsgFormat)) == 0 {
		tmp.ConsoleMsgFormat = DEFAULT_MSG_FORMAT
	}

	// email log default
	if stringLevel[tmp.EmailLevel] == 0{
		//fmt.Fprintln(os.Stdout,"unsupport level for email: "+tmp.EmailLevel)
		tmp.EmailLevel = "Warn"
	}
	if len(strings.TrimSpace(tmp.EmailMsgFormat)) == 0 {
		tmp.EmailMsgFormat = DEFAULT_MSG_FORMAT
	}

	// remote log default
	if stringLevel[tmp.RemoteLevel] == 0{
		//fmt.Fprintln(os.Stdout,"unsupport level for remote: "+tmp.RemoteLevel)
		tmp.RemoteLevel = "Warn"
	}
	if len(strings.TrimSpace(tmp.RemoteMsgFormat)) == 0 {
		tmp.RemoteMsgFormat = DEFAULT_MSG_FORMAT
	}

	cfg = tmp

	// when config success, restart all output writer
	registerLoggers()

	return nil
}


// generate the log struct and push to logChan, then wait to be write
func write(level int,msg... interface{})  {

	_,file,line,ok := runtime.Caller(2)
	if !ok {
		file = "???"
		line = 0
	}
	_,filename := path.Split(file)


	lg := logPool.Get().(*log)
	lg.Time = time.Now()
	lg.Msg = msg
	lg.Level = level
	lg.FileName = filename
	lg.Line = line

	lock.Lock()
	logChan <- lg
	lock.Unlock()
}

func SetFileLevel(level string) error {
	if stringLevel[level] != 0{
		cfg.FileLevel = level
		return nil
	}
	return errors.New("unsupport level:"+level)
}

func SetConsoleLevel(level string) error {
	if stringLevel[level] != 0{
		cfg.ConsoleLevel = level
		return nil
	}
	return errors.New("unsupport level:"+level)
}

func Trace(msg ...interface{})  {
	write(levelTrace,msg)
}

func Tracef(format string, a ...interface{})  {
	msg := []interface{} {fmt.Sprintf(format, a...)}
	write(levelTrace,msg)
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

func Critical(msg ...interface{})  {
	write(levelCritical,msg)
}

func Criticalf(format string, a ...interface{})  {
	msg := []interface{} {fmt.Sprintf(format, a...)}
	write(levelCritical,msg)
}


// write logs when logChan is not empty
// stop log when receive signal "stop"
func startLogger()  {
	for{
		select {
		case lg := <- logChan:
			write2outputs(*lg)
			logPool.Put(lg)
		case sg := <- signalChan:
			flush()
			if sg == "close"{
				for key,_ := range outputs{
					outputs[key].Close()
				}
				outputs = nil
				stop = true
			}
		}

		if stop{
			break
		}
	}
}

func write2outputs(lg log) {
	for key,_ := range outputs{
		outputs[key].Write(lg)
	}
}


func flush() {

}


// interface of log ouput
type outPut interface {
	Set() error       // set necessary info for output
	Write(lg log) error    // written to a specific output
	Close()          // close the writer
	Flush()
}

// delete log out put by name
func delLog(name string)  {
	out,ok := outputs[name].(outPut)
	if ok{
		out.Close()
		fmt.Println("output deleted:",name)
	}
}

// generate the full log message from log struct
func genLogMsg(formater string, lg log) string {
	fmt_msg := formater
	fmt_msg = strings.Replace(fmt_msg, "%Time",lg.Time.Format("20060102-15:04:05"),-1)
	fmt_msg = strings.Replace(fmt_msg, "%Level",levelString[lg.Level],-1)
	fmt_msg = strings.Replace(fmt_msg, "%Line",fmt.Sprint(lg.Line),-1)
	fmt_msg = strings.Replace(fmt_msg, "%File",lg.FileName,-1)

	msgs := lg.Msg[0].([]interface{})
	msg_str := ""
	for k,_ := range msgs {
		msg_str = msg_str +fmt.Sprintf("%+v ",msgs[k])
	}
	return strings.Replace(fmt_msg, "%Msg",msg_str,-1) + "\n"
}




