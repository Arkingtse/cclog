package main

import (
	log "tes/cclog"
	"time"
	"fmt"
)

type kk struct  {
	Name string
	Age int
}

func main()  {

	k := kk{"shit", 12}

	fmt.Println("--------------before load config.json---------------",time.Now())
	log.Trace("tes trace","nothing",12,k)
	log.Info("tes info")
	log.Debug("tes debug")
	log.Warn("tes warn")
	log.Error("tes error")
	log.Critical("tes critical")

	time.Sleep(1 * time.Second)
	if err:=log.ConfigFromFile("config.json");err !=nil{
		fmt.Println("fialed to load config.json:",err)
	}
	fmt.Println("--------------after load config.json----------------",time.Now())


	log.Trace("tes trace","nothing",12,k)
	log.Info("tes-----------info")
	log.Debug("tes-----------debug")
	log.Warn("tes-----------warn")
	log.Error("tes-----------error")
	log.Critical("tes-----------critical")

	time.Sleep(1 * time.Second)
	fmt.Println("--------------tes formater--------------------------",time.Now())

	log.Tracef("%v--%v--%v","Tracef",1, true)
	log.Infof("%v--%v--%v","Infof",1, true)
	log.Debugf("%v--%v--%v","Debugf",1, true)
	log.Warnf("%v--%v--%v","Warnf",1, true)
	log.Errorf("%v--%v--%v","Errorf",1, true)
	log.Criticalf("%v--%v--%v","Criticalf",1, true)

	// wait all log to be print
	time.Sleep(1 * time.Second)

}