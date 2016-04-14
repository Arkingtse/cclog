package cclog

import (
	"sync"
	"os"
	"fmt"
	"time"
	"path/filepath"
)

type fileLog struct {
	sync.Mutex
	writer *os.File
	fileName string

	curLine int
	curSize int
}

func (w *fileLog)Write(lg log) error {
	if lg.Level < stringLevel[cfg.FileLevel]{
		return nil
	}

	fmt_msg := genLogMsg(cfg.FileMsgFormat,lg)

	if w.needRotate(){
		w.Lock() // block write when rotate
		if err := w.doRotate();err !=nil{
			fmt.Fprintf(os.Stderr,"logfile(%q): %v\n",w.fileName,err)
		}
		w.Unlock()
	}


	w.Lock() // block rotate when write
	_,err := w.writer.Write([]byte(fmt_msg))
	if err == nil {
		//w.curLine ++
		//w.curSize = w.curSize + len(fmt_msg)
	}
	w.Unlock()

	return err
}

func (w *fileLog)Set() error {

	if cfg.FileRotateType == "line" || cfg.FileRotateType == "size" {
		//w.fileName = time.Now().Format("200601021504")
	}else {
		// default rotate daily
		w.fileName = time.Now().Format(cfg.FileNameFormat)
	}


	if err := w.noFileThenCreate(w.fileName);err!=nil{
		return err
	}

	w.startLogger()  // start the file writer

	return nil
}

func (w *fileLog)noFileThenCreate(name string) error {
	folder,_ := filepath.Split(name)

	if 0 != len(folder) {
		err := os.MkdirAll(folder, 0766)
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
	return os.OpenFile(w.fileName, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0660)
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

	if cfg.FileRotateType == "daily"{
		return time.Now().Format(cfg.FileNameFormat) != w.fileName
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
	w.fileName = time.Now().Format(cfg.FileNameFormat)

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