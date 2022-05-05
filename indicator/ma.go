package indicator

import (
	"fmt"
	"math"
)

type Ma struct {
	n1 int
	n2 int
	n3 int
}

// NewMa(5)
func NewMa(n1 int) *Ma {
	return &Ma{n1: n1}
}

func (this *Ma) WithMa(bids []*Kline) (ma []float64) {
	l := len(bids)
	ma = make([]float64, l)
	for i := 0; i < l; i++ {
		if i < this.n1 {
			continue
		}
		fmt.Println("aaa", bids[i].Close, bids[i].K)
		total := 0.0
		// @todo 这里错了，记得改
		for _, v := range bids[i-this.n1 : i] {
			total += v.Close
		}
		ma[i] = math.Round(total/float64(i)*1000) / 1000
	}
	fmt.Println(ma)
	return
}
