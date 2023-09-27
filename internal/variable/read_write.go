package variable

import (
	"encoding/json"
	"sync"

	"github.com/scutrobotlab/asuwave/pkg/jsonfile"
)

type Mod int

const (
	RD Mod = iota
	WR
)

type T struct {
	Board      uint8
	Name       string
	Type       string
	Addr       uint32
	Data       float64
	Tick       uint32
	Inputcolor string
	SignalGain float64
	SignalBias float64
}

type RWMap struct { // 一个读写锁保护的线程安全的map
	sync.RWMutex // 读写锁保护下面的map字段
	m            map[uint32]T
}

var to []RWMap = []RWMap{{
	m: make(map[uint32]T, 0),
}, {
	m: make(map[uint32]T, 0),
}}

func SetAll(o Mod, v map[uint32]T) {
	to[o].Lock() // 锁保护
	defer to[o].Unlock()
	to[o].m = v
	jsonfile.Save(jsonPath[o], to[o].m)
}

// 以json格式获取所有Mod变量
func GetAll(o Mod) ([]byte, error) {
	to[o].RLock() // 锁保护
	defer to[o].RUnlock()
	return json.Marshal(to[o].m)
}

func GetKeys(o Mod) (keys []uint32) {
	to[o].RLock() // 锁保护
	defer to[o].RUnlock()
	for k := range to[o].m {
		keys = append(keys, k)
	}
	return
}

// Get 从map中读取一个值
// o 是 Mod 类型的参数，表示map的模块
// k 是 uint32 类型的参数，表示要查找的键
// 返回值有两个：T 类型的值和一个bool类型，表示该键是否存在
func Get(o Mod, k uint32) (T, bool) {
	to[o].RLock()            // 为读取操作加上读锁
	defer to[o].RUnlock()    // 函数执行完毕后释放读锁
	v, existed := to[o].m[k] // 在锁的保护下从map中读取值
	return v, existed        // 返回读取到的值和是否存在的bool值
}

// Set 设置一个键值对到map中
// o 是 Mod 类型的参数，表示map的模块
// k 是 uint32 类型的参数，表示要设置的键
// v 是 T 类型的参数，表示要设置的值
func Set(o Mod, k uint32, v T) {
	to[o].Lock()                        // 为写入操作加上写锁
	defer to[o].Unlock()                // 函数执行完毕后释放写锁
	to[o].m[k] = v                      // 设置键值对到map中
	jsonfile.Save(jsonPath[o], to[o].m) // 将map的内容保存到json文件中
}

// Delete 从map中删除一个键
// o 是 Mod 类型的参数，表示map的模块
// k 是 uint32 类型的参数，表示要删除的键
func Delete(o Mod, k uint32) {
	to[o].Lock()                        // 为删除操作加上写锁
	defer to[o].Unlock()                // 函数执行完毕后释放写锁
	delete(to[o].m, k)                  // 从map中删除键
	jsonfile.Save(jsonPath[o], to[o].m) // 将map的剩余内容保存到json文件中
}
