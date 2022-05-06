package indicator

import (
	"github.com/LogicHou/bftr/utils"
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
		total := 0.0
		j := i + 1
		for _, v := range bids[j-this.n1 : j] {
			total += v.Close
		}
		ma[i] = utils.FRound2(total / float64(this.n1))
	}
	return
}

func (this *Ma) CurrentMa(bids []*Kline, curClose float64) float64 {
	total := curClose
	for _, v := range bids[len(bids)-(this.n1-1):] {
		total += v.Close
	}
	ma := utils.FRound2(total / float64(this.n1))
	return ma
}
