package getd

import (
	"testing"
	"github.com/carusyte/stock/model"
	"github.com/sirupsen/logrus"
)

func TestFinMark(t *testing.T) {
	s := &model.Stock{}
	s.Code = "601377"
	s.Name = "兴业证券"
	ss := new(model.Stocks)
	ss.Add(s)
	finMark(ss)
}

func TestCalcIndics(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	stks := StocksDb()
	allstk := new(model.Stocks)
	for _, s := range stks {
		allstk.Add(s)
	}
	CalcIndics(allstk)
}