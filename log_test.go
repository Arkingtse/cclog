package cclog

import (
	"testing"
	rd "github.com/Pallinder/go-randomdata"
	"math/rand"
)

func TestLog(t *testing.T)  {
	Trace("0 --> tes trace")
	Info("1 --> tes info")
	Debug("2 --> tes debug")
	Warn("3 --> tes warn")
	Error("4 --> tes error")
	Critical("5 --> tes critical")

}

func BenchmarkLog(b *testing.B)  {
	// 每次写入12条log    12*10000   2.6s

	for i := 0;i <b.N;i++{
		Trace("0 --> tes trace",rd.Letters(rand.Intn(100)))
		Info("1 --> tes info",rd.Letters(rand.Intn(100)))
		Debug("2 --> tes debug",rd.Letters(rand.Intn(100)))
		Warn("3 --> tes warn",rd.Letters(rand.Intn(100)))
		Error("4 --> tes error",rd.Letters(rand.Intn(100)))
		Critical("5 --> tes critical",rd.Letters(rand.Intn(100)))

		Tracef("%v--%v--%v--%v","Tracef",1, true,rd.Letters(rand.Intn(100)))
		Infof("%v--%v--%v--%v","Infof",1, true,rd.Letters(rand.Intn(100)))
		Debugf("%v--%v--%v--%v","Debugf",1, true,rd.Letters(rand.Intn(100)))
		Warnf("%v--%v--%v--%v","Warnf",1, true,rd.Letters(rand.Intn(100)))
		Errorf("%v--%v--%v--%v","Errorf",1, true,rd.Letters(rand.Intn(100)))
		Criticalf("%v--%v--%v--%v","Criticalf",1, true,rd.Letters(rand.Intn(100)))
	}
}