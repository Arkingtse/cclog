package cclog

import (
	"io"
	"sync"
	"runtime"
	"strings"
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
	newBrush("1;34"), // Trace              blue
	newBrush("1;34"), // Debug              blue
	newBrush("1;34"), // Info               blue
	newBrush("1;33"), // Warn               yellow
	newBrush("1;31"), // Error              red
	newBrush("1;35"), // Critical           magenta
}

type consoleLog struct {
	sync.Mutex
	writer io.Writer

	enable bool
	level int
	colorful bool

	msgFormater string
}

func (o *consoleLog)Set() error {
	o.colorful = true
	if runtime.GOOS == "windows"{  // windows 平台不支持带颜色的log
		o.colorful = false
	}

	o.level = stringLevel[cfg.ConsoleLevel]

	o.msgFormater = cfg.FileMsgFormat
	if strings.TrimSpace(o.msgFormater) == ""{
		o.msgFormater = "%Time [%Level] %Msg --[%Line]%File"  // 默认消息格式
	}

	o.writer = os.Stdout

	return nil
}

func (o *consoleLog)Write(lg log) error {
	if lg.Level < o.level {
		return nil
	}

	msg := genLogMsg(o.msgFormater, lg)

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