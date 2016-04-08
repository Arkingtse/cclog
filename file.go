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
	nameFormater string  // file name format
	msgFormater string  // log msg format

	rotate bool    // need rotate
	rotateType string  // rotate type

	maxRotate int64 // max files to be saved
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


	w.Lock() //
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
		w.maxRotate = 20   // default save 20 file
	}

	w.filePerm = 0660

	w.nameFormater = cfg.FileNameFormat
	if strings.TrimSpace(w.nameFormater) == ""{  // set default file name
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
	default:  // default rotate daily
		w.rotateType = "daily"
		w.fileName = time.Now().Format(cfg.FileNameFormat)
	}


	if err := w.noFileThenCreate(w.fileName);err!=nil{
		return err
	}

	w.msgFormater = cfg.FileMsgFormat
	if strings.TrimSpace(w.msgFormater) == ""{
		w.msgFormater = "%Time [%Level] %Msg --[%Line]%File"  // default msg format
	}


	w.level = stringLevel[cfg.FileLevel]
	w.rotate = cfg.FileRotate

	w.startLogger()  // start the file writer

	return nil
}

func (w *fileLog)noFileThenCreate(name string) error {
	folder,_ := filepath.Split(name)

	if 0 != len(folder) {
		err := os.MkdirAll(folder, 0660)
		if err != nil {
			return err
		}
	}
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
		fmt.Println("error when create file:", err)
		return err
	}
	if w.writer != nil{
		w.writer.Close()
	}

	w.writer = file
	return w.initFd()
}

func (w *fileLog)needRotate() bool {

	// if file not exist, then need rotate
	if _,err := os.Stat(w.fileName);err!=nil && !os.IsExist(err){
		return true
	}

	if w.rotateType == "daily"{
		return time.Now().Format(w.nameFormater) != w.fileName
	}else {
		//return w.curLine >= w.maxLines || w.curSize >= w.maxSize
	}
	return false
}

func (w *fileLog)doRotate() error {
	if err := w.noFileThenCreate(w.fileName);err!=nil{
		return err
	}

	// file exist
	w.writer.Close()
	w.fileName = time.Now().Format(w.nameFormater)

	if err := w.startLogger();err !=nil{
		return fmt.Errorf("error when rotate file: %\n", err)
	}

	return nil
}

func newFileLog() *fileLog {
	log := new(fileLog)
	log.Set()
	return log
}