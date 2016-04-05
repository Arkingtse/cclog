package main

import (
	log "tes/cclog"
	"time"
	"fmt"
	"os"
	"path/filepath"
)

type kk struct  {
	Name string
	Age int
}

func main()  {

	k := kk{"shit", 12}
	log.Trace("kksksksk","nothing",12,k)
	log.Info("tes info")
	log.Debug("tes debug")
	log.Warn("tes warn")
	log.Error("tes error")
	log.Critical("tes critical")

	//time.Sleep(1 * time.Second)

	fmt.Println("-----------------------------")
	wd,e1 := filepath.Abs(filepath.Dir(os.Args[0]))
	fmt.Println("当前路径:",wd,e1)

	if err:=log.ConfigFromFile("/Users/akc/go/src/tes/cctes/log.json");err !=nil{
		fmt.Println("加载配置失败:",err)
	}


	log.Trace("kksksksk","nothing",12,k)
	log.Info("tes-----------info")
	log.Debug("tes-----------debug")
	log.Warn("tes-----------warn")
	log.Error("tes-----------error")
	log.Critical("tes-----------critical")

	fmt.Println("-----------------------------end")

	log.Tracef("%v--%v--%v","Tracef",1, true)
	log.Infof("%v--%v--%v","Infof",1, true)
	log.Debugf("%v--%v--%v","Debugf",1, true)
	log.Warnf("%v--%v--%v","Warnf",1, true)
	log.Errorf("%v--%v--%v","Errorf",1, true)
	log.Criticalf("%v--%v--%v","Criticalf",1, true)

	time.Sleep(1 * time.Second)
}