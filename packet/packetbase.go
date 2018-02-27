package packet

import (
	"bytes"
	b "encoding/binary"
	"fmt"
)

const PKG_HEADER_LEN = 32
const PKG_HEADER_KEY = 16

//包头的设计使得在不解包的情况下能够对数据包进行排队，路由和查找，包头总长度是32字节
type PkgHeader struct {
	PkgLen    int32                //消息包的总长度=PKG_HEADER_LEN+PkgBodyLen, 包括了包头和消息包的长度
	ProductID int32                //productid
	CMD       int32                //命令号
	CliSeq    int32                //客户端的包序列号
	HeadKey   [PKG_HEADER_KEY]byte //16字节的key
}

func (p *PkgHeader) String() string {
	return fmt.Sprintf("pkglen:%d productid:%d cmd:%d cliseq:%d headkey:%s", p.PkgLen, p.ProductID, p.CMD, p.CliSeq, p.HeadKey)
}

type Packet struct {
	Header PkgHeader
	Body   []byte //body的二进制数据
}

func (c *Packet) String() string {
	return fmt.Sprintf("pkglen:%d productid:%d cmd:%d cliseq:%d headkey:%s", c.Header.PkgLen, c.Header.ProductID, c.Header.CMD, c.Header.CliSeq, c.Header.HeadKey)
}

func SerializeHeader(header PkgHeader) *bytes.Buffer {
	sbyte := make([]byte, 0, PKG_HEADER_LEN)
	buff := bytes.NewBuffer(sbyte)
	b.Write(buff, b.BigEndian, int32(header.PkgLen))
	b.Write(buff, b.BigEndian, header.ProductID)
	b.Write(buff, b.BigEndian, header.CMD)
	b.Write(buff, b.BigEndian, header.CliSeq)
	sbytet := header.HeadKey[0:]
	buff.Write(sbytet)
	return buff
}

func ParseHeader(r *bytes.Reader) (PkgHeader, error) {
	header := PkgHeader{}
	err := b.Read(r, b.BigEndian, &header.PkgLen)
	if nil != err {
		return header, err
	}

	err = b.Read(r, b.BigEndian, &header.ProductID)
	if nil != err {
		return header, err
	}

	err = b.Read(r, b.BigEndian, &header.CMD)
	if nil != err {
		return header, err
	}

	err = b.Read(r, b.BigEndian, &header.CliSeq)
	if nil != err {
		return header, err
	}

	err = b.Read(r, b.BigEndian, &header.HeadKey)
	if nil != err {
		return header, err
	}

	return header, err

}

func DecodeToPkg(buf []byte) *Packet {
	bytesbuff := bytes.NewReader(buf)
	pkgheader, err := ParseHeader(bytesbuff)
	if err != nil {
		return nil
	}

	return &Packet{
		Header: pkgheader,
		Body:   buf[PKG_HEADER_LEN:]}
}

/*
func NewPkgFromMsg(msg basesocial.Msg) *Packet{
	rspbuf, err := proto.Marshal(msg)
	if(err != nil){
		logs.Error("marshal msg faield.err:%s", err)
		return nil
	}

	h := PkgHeader{
		PkgLen:    int32(len(rspbuf) + PKG_HEADER_LEN),
		ProductID: msg.Head.ProductID,
		CMD:       CMD,
		CliSeq:    CliSeq,
		HeadKey:   headerkey}

	return &Packet{
		Header: h,
		Body:   body}
}*/

func NewPkg(ProductID int32, CMD int32, CliSeq int32, headerkey [16]byte, body []byte) *Packet {
	h := PkgHeader{
		PkgLen:    int32(len(body) + PKG_HEADER_LEN),
		ProductID: ProductID,
		CMD:       CMD,
		CliSeq:    CliSeq,
		HeadKey:   headerkey}

	return &Packet{
		Header: h,
		Body:   body}
}

func NewRspPkg(ProductID int32, CMD int32, CliSeq int32, headerkey [16]byte, body []byte) *Packet {
	h := PkgHeader{
		PkgLen:    int32(len(body) + PKG_HEADER_LEN),
		ProductID: ProductID,
		CMD:       CMD,
		CliSeq:    CliSeq,
		HeadKey:   headerkey}

	return &Packet{
		Header: h,
		Body:   body}
}

func (p *Packet) Serialize() []byte {
	buff := SerializeHeader(p.Header)
	buff.Write(p.Body)
	return buff.Bytes()
}
