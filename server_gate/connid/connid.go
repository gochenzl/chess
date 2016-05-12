package connid

import "sync"

const (
	InvalidId  = 0
	maxConnNum = 1000000 // 最大的连接个数
	maxConnId  = 3000000 // 最大id
)

var (
	usedConnNum   int // 已经使用的连接个数
	curIndex      int
	idPool        []uint32
	idRecyclePool []uint32
	idRecycleBits []uint64
	mu            sync.Mutex
)

func Init() {

	idPool = make([]uint32, maxConnId)
	idRecyclePool = make([]uint32, 0, maxConnId)
	idRecycleBits = make([]uint64, maxConnId/64+1)

	curIndex = 0

	var id uint32 = 1
	for i := 0; i < maxConnId; i++ {
		idPool[i] = id
		id++
	}
}

func Get() uint32 {
	mu.Lock()
	defer mu.Unlock()

	if usedConnNum >= maxConnNum {
		return InvalidId
	}

	if curIndex == len(idPool) {
		idPool, idRecyclePool = idRecyclePool, idPool[:0]
		curIndex = 0
		for i := 0; i < len(idRecycleBits); i++ {
			idRecycleBits[i] = 0
		}
	}

	id := idPool[curIndex]
	curIndex++
	usedConnNum++
	return id
}

func Release(id uint32) bool {
	if id == InvalidId || id > maxConnId {
		return false
	}

	mu.Lock()
	defer mu.Unlock()

	if idRecycleBits[id>>6]&(1<<(id&(64-1))) != 0 {
		return false
	}

	idRecyclePool = append(idRecyclePool, id)
	idRecycleBits[id>>6] |= 1 << (id & (64 - 1))

	if usedConnNum > 0 {
		usedConnNum--
	}

	return true
}

func Remain() int {
	mu.Lock()
	n := maxConnNum - usedConnNum
	mu.Unlock()
	return n
}
