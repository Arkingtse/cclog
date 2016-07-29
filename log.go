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
	"path/filepath"
)

const DEFAULT_MSG_FORMAT = "%Time [%Level] %Msg --[%Line]%File"
const DEFAULT_FILE_FORMAT = "log/20060102.log"

// Trace, Debug, Info, Warn, Error, Critical
const (
	//levelTrace = iota + 1   // Trace message
	levelDebug = iota + 1   // Debug message
	levelInfo               // Info message
	levelWarn               // Warn message
	levelError              // Error message
	levelFatal              // Fatal message(not)
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


type CMap struct {
	sync.Mutex
	itms map[string]outPut
}

func (o *CMap)Set(key string, value outPut) {
	o.Lock()
	o.itms[key] = value
	o.Unlock()
}

func (o *CMap)Get(key string) outPut {
	return o.itms[key]
}

func (o *CMap)Del(key string)  {
	o.Lock()
	delete(o.itms, key)
	o.Unlock()
}


func (o *CMap)CloseAll(){
	for key,_ := range o.itms{
		o.itms[key].Close()
	}
}

func (o *CMap)WriteAll(lg log){
	for key,_ := range o.itms{
		o.itms[key].Write(lg)
	}
}


func NewCMap() *CMap {
	return &CMap{itms:make(map[string]outPut)}
}




var lock sync.Mutex
var cfg config
var logChan = make(chan log, 10000)
var signalChan = make(chan string)
var outputs = NewCMap()

var stop = false
var enableFile = false
var zipLogFile = false

// 运行过程没有写操作是安全的
var levelString = make(map[int]string)
var stringLevel = make(map[string]int)
var rotateType  = make(map[string]int)


func init() {
	//levelString[levelTrace] = "Trace"
	levelString[levelDebug] = "DEBU"
	levelString[levelInfo] = "INFO"
	levelString[levelWarn] = "WARN"
	levelString[levelError] = "ERRO"
	levelString[levelFatal] = "FATA"

	//stringLevel["Trace"] = levelTrace
	stringLevel["DEBU"] = levelDebug
	stringLevel["INFO"] = levelInfo
	stringLevel["WARN"] = levelWarn
	stringLevel["ERRO"] = levelError
	stringLevel["FATA"] = levelFatal

	rotateType["daily"] = 1
	rotateType["size"] = 2
	rotateType["line"] = 3


	// console log default
	cfg.Console = true
	cfg.ConsoleLevel = "DEBU"
	cfg.ConsoleMsgFormat = DEFAULT_MSG_FORMAT

	// file log default
	cfg.FileNameFormat = DEFAULT_FILE_FORMAT
	cfg.FileLevel = "INFO"
	cfg.FileMsgFormat = DEFAULT_MSG_FORMAT
	cfg.FileRotateType = "daily"
	cfg.FileMaxRotate = 20
	cfg.FileRotate = true

	// email log default
	cfg.EmailLevel = "INFO"
	cfg.EmailMsgFormat = DEFAULT_MSG_FORMAT

	// remote log default
	cfg.RemoteLevel = "INFO"
	cfg.RemoteMsgFormat = DEFAULT_MSG_FORMAT

	// register all output
	registerLoggers()
	// start log writer server
	go startLogger()
}

func registerLoggers()  {
	lock.Lock()  // block all writer when update log writer

	outputs.CloseAll()

	if enableFile{
		outputs.Set("file", newFileLog())
	}
	outputs.Set("console", newConsoleLog())

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
	// set log file name to full path
	if !filepath.IsAbs(tmp.FileNameFormat){
		exePath,err := filepath.Abs(os.Args[0])
		if err != nil {
			return err
		}
		logDir := filepath.Dir(exePath)
		tmp.FileNameFormat = logDir+"/"+tmp.FileNameFormat
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


	lg := log{}
	lg.Time = time.Now()
	lg.Msg = msg
	lg.Level = level
	lg.FileName = filename
	lg.Line = line

	logChan <- lg
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

func EnableFile() {
	enableFile = true

	registerLoggers()
}

func ZipLog() {
	zipLogFile = true
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


// write logs when logChan is not empty
// stop log when receive signal "stop"
func startLogger()  {
	for{
		select {
		case lg := <- logChan:
			outputs.WriteAll(lg)
		case sg := <- signalChan:
			flush()
			if sg == "close"{
				outputs.CloseAll()
				outputs = nil
				stop = true
			}
		}

		if stop{
			break
		}
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




