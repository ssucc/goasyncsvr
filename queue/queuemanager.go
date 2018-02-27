package queue

/*
考虑多线程安全性?
*/

import (
	"container/list"
	. "github.com/fsvr/packet"
	"time"
)

type KeyList struct {
	num     int32
	pkgList *list.List
	stamp   int64 //秒数
}

func NewKeyList() *KeyList {
	pkgList := list.New()
	return &KeyList{
		num:     0,
		pkgList: pkgList}
}

func (k *KeyList) PushPacket(packet *Packet) {
	k.pkgList.PushBack(packet)
	k.num++
	k.stamp = int64(time.Now().Second())
}

func (k *KeyList) PopPacket() *Packet {
	if k.pkgList.Len() > 0 {
		pkg := k.pkgList.Front()
		k.pkgList.Remove(pkg)
		k.stamp = int64(time.Now().Second())
		k.num--
		return pkg.Value.(*Packet)
	}

	return nil
}

type QueueManager struct {
	keyQueue map[string]*KeyList
}

func NewQueueManager() *QueueManager {
	queuemanager := new(QueueManager)
	queuemanager.keyQueue = make(map[string]*KeyList)
	return queuemanager
}

//return true，表示是队列的第一个消息，return false， 不是队列的第一个消息
func (q *QueueManager) PushMsg(key string, pkg *Packet) bool {
	l, ok := q.keyQueue[key]
	if !ok {
		l = NewKeyList()
		q.keyQueue[key] = l
		return true
	}

	l.PushPacket(pkg)
	return false
}

/*
return 包的指针或者失败为nil
*/
func (q *QueueManager) PopMsg(key string) *Packet {
	l, ok := q.keyQueue[key]
	if !ok {
		return nil
	}

	pkg := l.PopPacket()
	if pkg == nil {
		delete(q.keyQueue, key)
		return nil
	}

	return pkg
}
