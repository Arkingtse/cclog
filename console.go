package cclog

import (
	"io"
	"sync"
	"runtime"
	"os"
)

type brush func(string) string

func newBrush(color string) brush {
	pre := "\033["
	reset := "\033[0m"
	return func(text string) string {
		return pre + color + "m" + text + reset
	}
}

var colors = []brush{
	newBrush("1;0"), // 0                   none
	//newBrush("1;34"), // Trace            blue
	newBrush("1;37"), // Debug              gray
	newBrush("1;34"), // Info               blue
	newBrush("1;33"), // Warn               yellow
	newBrush("1;31"), // Error              red
	newBrush("1;35"), // Fatal              magenta
}

type consoleLog struct {
	sync.Mutex
	writer io.Writer

	colorful bool
}

func (o *consoleLog)Set() error {
	o.colorful = true
	if runtime.GOOS == "windows"{  // windows 平台不支持带颜色的log
		o.colorful = false
	}

	o.writer = os.Stdout

	return nil
}

func (o *consoleLog)Write(lg log) error {
	if lg.Level < stringLevel[cfg.ConsoleLevel] {
		return nil
	}

	msg := genLogMsg(cfg.ConsoleMsgFormat, lg)

	if o.colorful{
		msg = colors[lg.Level](msg)
	}
	o.Lock()
	o.writer.Write([]byte(msg))
	o.Unlock()
	return nil
}

func (o *consoleLog)Close()  {

}

func (o *consoleLog)Flush()  {

}

func newConsoleLog() *consoleLog {
	log := new(consoleLog)
	log.Set()
	return log
}