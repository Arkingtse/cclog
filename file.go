package cclog

import (
	"sync"
	"os"
	"fmt"
	"time"
	"path/filepath"
	"strings"
	"io/ioutil"
	"archive/zip"
	"bytes"
	"errors"
)

type fileLog struct {
	sync.Mutex
	writer *os.File
	fileName string
	filePerm os.FileMode

	curLine int
	curSize int
}

func (w *fileLog)Write(lg log) error {
	if lg.Level < stringLevel[cfg.FileLevel]{
		return nil
	}

	fmt_msg := genLogMsg(cfg.FileMsgFormat,lg)

	if w.needRotate(){

		if err := w.doRotate();err !=nil{
			fmt.Fprintf(os.Stderr,"logfile(%q): %v\n",w.fileName,err)
		}

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

	w.filePerm = 0766


	if cfg.FileRotateType == "line" || cfg.FileRotateType == "size" {
		//w.fileName = time.Now().Format("200601021504")
	}else {
		// default rotate daily
		w.fileName = time.Now().Format(cfg.FileNameFormat)

		if !filepath.IsAbs(w.fileName){
			exePath,err := filepath.Abs(os.Args[0])
			if err != nil {
				return err
			}
			logDir := filepath.Dir(exePath)
			w.fileName = logDir+"/"+w.fileName
			cfg.FileNameFormat = logDir+"/"+cfg.FileNameFormat
		}
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
		err := os.MkdirAll(folder, w.filePerm)
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

	//// zip all logs except the current log file
	//go w.zipLog(w.fileName)

	return os.OpenFile(w.fileName, os.O_WRONLY|os.O_APPEND|os.O_CREATE, w.filePerm)
}

func (w *fileLog)initFd() error {
	finfo,err := w.writer.Stat()
	if err != nil {
		return errors.New("get stat err: %s"+err.Error())
	}
	w.curSize = int(finfo.Size())

	// 先不处理行数
	return nil
}

func (w *fileLog)startLogger() error {
	file,err := w.createFile()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error when create file:", err)
		return err
	}

	// 切换writer时锁定
	w.Lock()

	if w.writer != nil{
		w.writer.Close()
	}

	w.writer = file

	w.Unlock()


	// zip all logs except the current log file
	go w.zipLog(w.fileName)

	return w.initFd()
}

func (w *fileLog)needRotate() bool {

	// if file not exist, then need rotate
	if _,err := os.Stat(w.fileName);err!=nil && os.IsNotExist(err){
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

	w.fileName = time.Now().Format(cfg.FileNameFormat)
	if !filepath.IsAbs(w.fileName){
		exePath,err := filepath.Abs(os.Args[0])
		if err != nil {
			return err
		}
		logDir := filepath.Dir(exePath)
		w.fileName = logDir+"/"+w.fileName
	}

	if err := w.noFileThenCreate(w.fileName);err!=nil{
		return err
	}


	if err := w.startLogger();err !=nil{
		return errors.New("error when rotate file: "+err.Error())
	}

	return nil
}

func (w *fileLog)zipLog(nozipfile string) {
	logdir := filepath.Dir(nozipfile)


	filepath.Walk(logdir, func(path string, info os.FileInfo, err error) error {

		if !strings.HasSuffix(path, ".log") {
			// if not .log then return
			return nil
		}

		if time.Since(info.ModTime()) < 24 * time.Hour{
			return nil
		}


		if nozipfile == path{
			// not zip the new created file
			return nil
		}

		body, err := ioutil.ReadFile(path)
		if err != nil {
			// if read file err then return
			return nil
		}

		head, err := zip.FileInfoHeader(info)
		if err != nil {
			// if read file header err then return
			return nil
		}
		head.Method = 8 // 设定压缩算法

		dst, err := os.Create(path[:len(path) - 4] + ".zip")
		if err != nil {
			// if create zip fail then return
			return nil
		}
		defer dst.Close()

		buf := new(bytes.Buffer)
		defer buf.WriteTo(dst)

		zf := zip.NewWriter(buf)
		defer zf.Close()

		// write file to zip writer
		w, err := zf.CreateHeader(head)
		if err != nil {
			return nil
		}
		_,err = w.Write(body)
		if err==nil{
			// if write successful then remove the src file
			os.Remove(path)
		}


		return nil
	})
}


func newFileLog() *fileLog {
	log := new(fileLog)
	log.Set()
	return log
}