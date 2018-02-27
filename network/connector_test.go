package network

import (
	. "github.com/smartystreets/goconvey"
	"testing"
)

func TestRead(t *testing.T) {
	Convey("发送数据", t, func() {
		So(Add(1, 2), ShouldEqual, 3)

	})
}


func TestWrite(t *testing.T){
	Convey("接受数据", t, func(){
		Convey("case 1", func(){

			})
		//里面不需要传入t参数，否则会panic
		Convey("case 2", func*(){

			})
		})

}


func ReadByte(){
	
	
}
