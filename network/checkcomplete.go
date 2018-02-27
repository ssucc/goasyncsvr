package network

import (
	b "encoding/binary"
	"github.com/astaxie/beego/logs"
)

const (
	//最大packet的字节数
	MAX_PACKET_BYTES int = 8 * 1024 * 1024
	PACKET_HEAD_LEN  int = 32 //请求头部长度	 int
)

/*
	protocol of tcp connect.
	return 0, no a complete pkg, need read more, error is nil
	return >0, means the whole pkglen, error is nil
	return <0, means error pkt, error is 1, the connection need be closed
*/
func CheckComplete(buf []byte, length int) int {
	if length < PACKET_HEAD_LEN {
		return 0
	}

	headlen := int(b.BigEndian.Uint32(buf))
	logs.Info("headlen:%d", headlen)
	if headlen > MAX_PACKET_BYTES {
		return -1
	}

	if headlen > length {
		return 0
	}

	return headlen
}
