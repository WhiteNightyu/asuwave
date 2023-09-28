package variable

import (
	"github.com/golang/glog"
	"github.com/scutrobotlab/asuwave/pkg/slip"
)

type CmdT struct {
	Board  uint8
	Length int
	Addr   uint32
	Tick   uint32
	Data   [8]byte
}

// Unpack 在如山的信笺里，找寻变量的回音
func Unpack(data []byte) ([]CmdT, []byte) {

	// 分开信笺
	ends := []int{}
	for i, d := range data {
		if d == slip.END {
			ends = append(ends, i)
		}
	}

	if len(ends) == 0 { // 若是无言
		return []CmdT{}, []byte{}
	}

	vars := []CmdT{}

	// 探寻每一封信
	for i := 1; i < len(ends); i++ {
		// 此信浅薄，难载深情
		if ends[i]-ends[i-1] < 20 {
			continue
		}
		// 解开此信
		glog.Infoln(data[ends[i-1] : ends[i]+1])
		pack, err := slip.Unpack(data[ends[i-1] : ends[i]+1])
		if err != nil {
			glog.Errorln(err.Error())
			continue
		}
		glog.Infoln(pack)
		// 非变量之回音，或无合法之落款，则弃之
		if len(pack) != 20 || ActMode(pack[1]) != SubscribeReturn || pack[19] != '\n' {
			glog.V(1).Infoln("Not Subscribereturn pack", pack)
			continue
		}
		// 聆听变量的回音
		v := CmdT{
			Board:  pack[0],
			Length: int(pack[2]),
			Addr:   BytesToUint32(pack[3:7]),
			Data:   *(*[8]byte)(pack[7:15]),
			Tick:   BytesToUint32(pack[15:19]),
		}
		// 加入变量列表
		vars = append(vars, v)
	}
	f := ends[len(ends)-1]               // 最后的结束，
	newbuff := make([]byte, len(data)-f) // 也是新的开始。
	if len(newbuff) == 0 {
		return vars, newbuff
	}
	copy(newbuff, data[f:]) // 留存残篇。
	return vars, newbuff    // 变量的回音，仍有余音
}

// Filt 从茫茫 vars 中，寻找我所挂念的 to[RD] ，记录在列表 chart 中。
// 所有的 add 我都难以忘记，所有的 del 我都不愿提起
func Filt(vars []CmdT) (chart []ChartT, add []CmdT, del []CmdT) {
	to[RD].RLock()
	defer to[RD].RUnlock()

	chart = []ChartT{}
	add = []CmdT{} // 有些变量，我难以忘记
	del = []CmdT{} // 有些变量，我不愿提起

	addrs := map[uint32]bool{}

	for _, v := range vars {
		// 它是我要找的那个变量吗？
		if r, ok := to[RD].m[v.Addr]; ok { // 是的，我还挂念着它
			r.Tick = v.Tick
			r.Data = SpecFromBytes(r.Type, v.Data[:])
			chart = append(chart, ChartT{
				Board: r.Board,
				Name:  r.Name,
				Data:  r.SignalGain*r.Data + r.SignalBias,
				Tick:  r.Tick,
			})
		} else { // 不是的，请忘了它
			del = append(del, v)
		}
	}

	// 我所挂念的，它们都还在吗
	for _, r := range to[RD].m {
		if _, ok := addrs[r.Addr]; !ok {
			// 我很想它，下次请别忘记
			add = append(add, CmdT{
				Board:  r.Board,
				Length: TypeLen[r.Type],
				Addr:   r.Addr,
			})
		}
	}
	return
}
