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


// Trace, Debug, Info, Warn, Error, Critical
const (
	levelTrace = iota    // Trace message
	levelDebug           // Debug message
	levelInfo            // Info message
	levelWarn            // Warn message
	levelError           // Error message
	levelCritical        // Critical message(not)
)


// log config
type config struct {

	// 输出到文件,默认输出到文件
	FileLevel string  `json:"fileLevel"`
	FileNameFormat string  `json:"fileNameFormater"`
	FileMsgFormat string  `json:"fileMsgFormater"`
	FileRotate bool  `json:"fileRotate"`   // 是否切换(切割)日志文件
	FileRotateType string  `json:"fileRotateType"`   // 分割方式, 文件大小/日期/ size/date
	FileMaxRotate int64  `json:"fileMaxRotate"`   // 日志保存数量
	FileMaxLine int  `json:"fileMaxLine"`  // 最大行数
	FileMaxSize int  `json:"fileMaxSize"`  // 最大文件大小


	// 输出到控制台
	Console bool  `json:"console"`
	ConsoleLevel string  `json:"consoleLevel"`
	ConsoleMsgFormat string  `json:"consoleMsgFormat"`

	// 发送邮件, 填了邮箱就发
	Email string  `json:"email"`
	EmailLevel string  `json:"emailLevel"`
	EmailMsgFormat string  `json:"emailMsgFormat"`

	// 输送远程, 填了地址就发
	Remote string  `json:"remote"`
	RemoteLevel string  `json:"remoteLevel"`
	RemoteMsgFormat string  `json:"remoteMsgFormat"`
}




// log信息结构
type log struct {
	Time time.Time
	Level int
	Msg []interface{}
	Line int
	FileName string
}

var logPool *sync.Pool


// 模块几个重要对象
var lock sync.Mutex
var cfg config
var logChan = make(chan *log, 10000)
var signalChan = make(chan string)
//var wg sync.WaitGroup  //等待所有输入写完?
var outputs = make(map[string]outPut)

var stop = false // log状态

var levelString = make(map[int]string)
var stringLevel = make(map[string]int)


func init() {  //默认输出
	levelString[0] = "Trace"
	levelString[1] = "Debug"
	levelString[2] = "Info"
	levelString[3] = "Warn"
	levelString[4] = "Error"
	levelString[5] = "Critical"

	stringLevel["Trace"] = 0
	stringLevel["Debug"] = 1
	stringLevel["Info"] = 2
	stringLevel["Warn"] = 3
	stringLevel["Error"] = 4
	stringLevel["Critical"] = 5

	cfg.Console = true
	cfg.ConsoleLevel = "Trace"

	// 默认文件名与消息格式
	cfg.FileMsgFormat = "%Time [%Level] %Msg --[%Line]%File"
	cfg.FileNameFormat = "log/20060102.log"
	cfg.FileLevel = "Trace"

	cfg.FileRotateType = "daily"

	// 初始化log对象池
	logPool = &sync.Pool{
		New:func() interface{}{
			return &log{}
		},
	}

	// 默认开文件log
	// 按配置注册其他log
	registerLoggers()
	// 开启协程后台写log到各输出
	go startLogger()
}

func registerLoggers()  {
	lock.Lock()  // 更换配置时锁定日志写入,更新完成后再允许写入日志

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

	switch tmp.FileLevel {
	case "Trace","Info","Debug","Warn","Error","Critical":
	default:
		fmt.Fprint(os.Stdout,"unsupport level for file: "+tmp.FileLevel)
		tmp.FileLevel = "Warn"
	}

	switch tmp.ConsoleLevel {
	case "Trace","Info","Debug","Warn","Error","Critical":
	default:
		fmt.Fprint(os.Stdout,"unsupport level for console: "+tmp.FileLevel)
		tmp.ConsoleLevel = "Warn"
	}

	cfg = tmp

	// 配置加载成功后
	registerLoggers()

	return nil
}


// 设置单条log内容
func write(level int,msg... interface{})  {

	_,file,line,ok := runtime.Caller(2)
	if !ok {
		file = "???"
		line = 0
	}
	_,filename := path.Split(file)

	// 获取一个现有的log对象,设置好后放到待写logChan中
	lg := logPool.Get().(*log)
	lg.Time = time.Now()
	lg.Msg = msg
	lg.Level = level
	lg.FileName = filename
	lg.Line = line

	lock.Lock()  // 当log被锁定时, 阻塞日志写入
	logChan <- lg
	lock.Unlock()
}

func SetFileLevel(level string) error {
	switch level {
	case "Trace", "Debug", "Info", "Warn", "Error", "Critical":
		cfg.FileLevel = level
		registerLoggers()
		return nil
	default:
		return errors.New("unsupport level")
	}
}

func SetConsoleLevel(level string) error {
	switch level {
	case "Trace", "Debug", "Info", "Warn", "Error", "Critical":
		cfg.ConsoleLevel = level
		registerLoggers()
		return nil
	default:
		return errors.New("unsupport level")
	}
}

func Trace(msg ...interface{})  {
	write(levelTrace,msg)
}

func Tracef(format string, a ...interface{})  {
	submsg := []interface{} {fmt.Sprintf(format, a...)}
	msg := []interface{} {submsg}
	write(levelTrace,msg)
}

func Debug(msg ...interface{})  {
	write(levelDebug,msg)
}

func Debugf(format string, a ...interface{})  {
	submsg := []interface{} {fmt.Sprintf(format, a...)}
	msg := []interface{} {submsg}
	write(levelDebug,msg)
}

func Info(msg ...interface{})  {
	write(levelInfo,msg)
}

func Infof(format string, a ...interface{})  {
	submsg := []interface{} {fmt.Sprintf(format, a...)}
	msg := []interface{} {submsg}
	write(levelInfo,msg)
}

func Warn(msg ...interface{})  {
	write(levelWarn,msg)
}

func Warnf(format string, a ...interface{})  {
	submsg := []interface{} {fmt.Sprintf(format, a...)}
	msg := []interface{} {submsg}
	write(levelWarn,msg)
}

func Error(msg ...interface{})  {
	write(levelError,msg)
}

func Errorf(format string, a ...interface{})  {
	submsg := []interface{} {fmt.Sprintf(format, a...)}
	msg := []interface{} {submsg}
	write(levelError,msg)
}

func Critical(msg ...interface{})  {
	write(levelCritical,msg)
}

func Criticalf(format string, a ...interface{})  {
	submsg := []interface{} {fmt.Sprintf(format, a...)}
	msg := []interface{} {submsg}
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


// 输出方式接口定义
type outPut interface {
	Set() error       // 设置日志格式
	Write(lg log) error    // 写入日志
	Close()          // 关闭日志
	Flush()          // 刷新日志
}

func delLog(name string)  {
	out,ok := outputs[name].(outPut)
	if ok{
		out.Close()
		fmt.Println("log关闭成功:",name)
	}
}

func genLogMsg(formater string, lg log) string {
	fmt_msg := formater
	fmt_msg = strings.Replace(fmt_msg, "%Time",lg.Time.Format("20060102-15:04:05"),-1)
	fmt_msg = strings.Replace(fmt_msg, "%Level",levelString[lg.Level],-1)
	fmt_msg = strings.Replace(fmt_msg, "%Line",fmt.Sprint(lg.Line),-1)
	fmt_msg = strings.Replace(fmt_msg, "%File",lg.FileName,-1)

	// default type of mulpara if [][]interface{}
	msgs := lg.Msg[0].([]interface{})
	msg_str := ""
	for k,_ := range msgs {
		msg_str = msg_str +fmt.Sprintf("%+v ",msgs[k])
	}

	return strings.Replace(fmt_msg, "%Msg",msg_str,-1) + "\n"
}




