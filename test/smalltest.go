package main

import (
	//"fmt"
	//"reflect"
	"flag"
	"github.com/fsvr/base"
	"github.com/astaxie/beego/logs"
)

type Foo struct {
}
type Bar struct {
	A string
	B int32
}

//用于保存实例化的结构体对象
var regStruct map[string]interface{}

func main() {
	/*
	type T struct {
		A int
		B string
	}

	t := T{23, "skidoo"}
	s := reflect.ValueOf(&t).Elem()
	typeOfT := s.Type()
	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		fmt.Printf("%d: %s %s = %v\n", i,
			typeOfT.Field(i).Name, f.Type(), f.Interface())
	}

	str := "Bar"
	if regStruct[str] != nil {
		t := reflect.ValueOf(regStruct[str]).Type()
		fmt.Printf("ttype:%s", t.Name())
		v := reflect.New(t).Elem()
		fmt.Println(v)
	}
	*/

	tfilename := flag.String("conf", "./cliet.conf", "-conf=./client.conf")
	flag.Parse()
	filename := *tfilename
	conf, err := base.NewConf("ini", filename)
	if err != nil {
		logs.Error("New Conf failed.err:%s", err)
		return
	}

	m := conf.GetSection("bidlist")
	for k,v := range(m){
		logs.Info("k:%s v:%s", k, v)
	}

	logs.Info("host:%s path1:%s", conf.GetFieldStr("host", "localhost"), conf.GetFieldStr("path1", "nopath"));
}

func init() {
	regStruct = make(map[string]interface{})
	regStruct["Foo"] = Foo{}
	regStruct["Bar"] = Bar{A: "testa", B: 1234}
}
