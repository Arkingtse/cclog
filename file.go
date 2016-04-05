package cclog

import (
	"sync"
	"os"
	"fmt"
	"strings"
	"time"
	"path/filepath"
)

type fileLog struct {
	sync.Mutex
	writer *os.File
	fileName string
	filePerm os.FileMode

	level int
	nameFormater string  // 文件名格式
	msgFormater string  // 日志格式

	rotate bool    // 是否切割日志
	rotateType string  // 切割方式

	maxRotate int64 // 保存最大天数
	maxLines int
	maxSize int

	curLine int
	curSize int
}

func (w *fileLog)Write(lg log) error {
	if lg.Level < w.level{
		return nil
	}

	fmt_msg := genLogMsg(w.msgFormater,lg)

	if w.needRotate(){
		w.Lock()
		if err := w.doRotate();err !=nil{
			fmt.Fprintf(os.Stderr,"logfile(%q): %v\n",w.fileName,err)
		}
		w.Unlock()
	}


	w.Lock() // 锁定文件写入
	_,err := w.writer.Write([]byte(fmt_msg))
	if err == nil {
		//w.curLine ++
		//w.curSize = w.curSize + len(fmt_msg)
	}
	w.Unlock()

	return err
}

func (w *fileLog)Set() error {

	w.maxRotate = cfg.FileMaxRotate
	if w.maxRotate == 0{
		w.maxRotate = 20   // 默认保存20个日志文件
	}

	w.filePerm = 0660

	w.nameFormater = cfg.FileNameFormat
	if strings.TrimSpace(w.nameFormater) == ""{  // 默认文件名
		if w.rotateType == "daily"{
			w.nameFormater = "log/20060102.log"
		}else {
			w.nameFormater = "log/1.log"
		}
	}

	switch cfg.FileRotateType {
	case "line","size":
		w.rotateType = cfg.FileRotateType
		w.fileName = cfg.FileNameFormat
	default:  // 默认按天拆分log
		w.rotateType = "daily"
		w.fileName = time.Now().Format(cfg.FileNameFormat)
	}


	folder,_ := filepath.Split(w.fileName)

	//fmt.Println("日志路径: ", folder)
	// 创建目标路径
	if 0 != len(folder) {
		err := os.MkdirAll(folder, 0767)
		if err != nil {
			return err
		}
	}

	w.msgFormater = cfg.FileMsgFormat
	if strings.TrimSpace(w.msgFormater) == ""{
		w.msgFormater = "%Time [%Level] %Msg --[%Line]%File"  // 默认消息格式
	}


	w.level = stringLevel[cfg.FileLevel]
	w.rotate = cfg.FileRotate

	w.startLogger()  // 打开文件

	return nil
}

func (w *fileLog)Close()  {
	w.writer.Close()
}

func (w *fileLog)Flush()  {
	w.writer.Sync()
}

func (w *fileLog)createFile() (*os.File,error) {
	return os.OpenFile(w.fileName, os.O_WRONLY|os.O_APPEND|os.O_CREATE, w.filePerm)
}

func (w *fileLog)initFd() error {
	finfo,err := w.writer.Stat()
	if err != nil {
		return fmt.Errorf("get stat err: %s",err)
	}
	w.curSize = int(finfo.Size())

	// 先不处理行数
	return nil
}

func (w *fileLog)startLogger() error {
	file,err := w.createFile()
	if err != nil {
		fmt.Println("创建日志文件失败:", err)
		return err
	}
	if w.writer != nil{
		w.writer.Close()
	}

	w.writer = file
	return w.initFd()
}

func (w *fileLog)needRotate() bool {
	if w.rotateType == "daily"{
		return time.Now().Format(w.nameFormater) != w.fileName
	}else {
		//return w.curLine >= w.maxLines || w.curSize >= w.maxSize
	}
	return false
}

func (w *fileLog)doRotate() error {
	_,err := os.Lstat(w.fileName)
	if err != nil {
		return err
	}

	// file exist
	w.writer.Close()
	w.fileName = time.Now().Format(w.nameFormater)

	if err := w.startLogger();err !=nil{
		return fmt.Errorf("切换文件出错: %\n", err)
	}

	return nil
}

func newFileLog() *fileLog {
	log := new(fileLog)
	log.Set()
	return log
}