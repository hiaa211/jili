package binancecollector

import (
	"container/heap"
	"log"
	"sync"
	"time"
)

// record 是 priorityQueue 中的元素
type record struct {
	symbol string
	utc    int64
	id     int64
}

func newRecord(symbol string, utc, id int64) *record {
	return &record{
		symbol: symbol,
		utc:    utc,
		id:     id,
	}
}

var mu sync.RWMutex

func (rs *records) ethbtc() *record {
	mu.RLock()
	var res *record
	for _, r := range *rs {
		if r.symbol == "ETHBTC" {
			res = r
			break
		}
	}
	mu.RUnlock()
	if res == nil {
		log.Fatal("没有找到 ETHBTC")
	}
	return res
}

// all 记录了 symbols 的数量
var all int

func (rs *records) isDelayed() bool {
	mu.RLock()
	defer mu.RUnlock()
	return all-len(*rs) > 12
}

func (rs *records) first() (symbol string, utc, id int64) {
	mu.RLock()
	symbol = (*rs)[0].symbol
	id = (*rs)[0].id
	utc = (*rs)[0].utc
	mu.RUnlock()
	log.Printf("the first symbol: %s,\t ID: %12d, Time: %s", symbol, id, time.Unix(0, utc*1000000))
	return
}

func (rs *records) pop() *record {
	mu.Lock()
	res := heap.Pop(rs).(*record)
	mu.Unlock()
	return res
}

func (rs *records) push(r *record) {
	mu.Lock()
	heap.Push(rs, r)
	mu.Unlock()
}

func (rs *records) update(utc, id int64) {
	mu.Lock()
	(*rs)[0].utc = utc
	(*rs)[0].id = id
	heap.Fix(rs, 0)
	mu.Unlock()
}

func (rs *records) isUpdated() bool {
	mu.RLock()
	latest := time.Unix(0, 1000000*(*rs)[0].utc)
	mu.RUnlock()
	return isToday(latest)
}

func isToday(date time.Time) bool {
	now := time.Now()
	return date.Year() == now.Year() &&
		date.Month() == now.Month() &&
		date.Day() == now.Day()
}

// records implements heap.Interface and holds entries.
type records []*record

func newRecords() *records {
	symbols := allSymbols()
	all = len(symbols)
	res := make(records, 0, len(symbols))
	for _, s := range symbols {
		tp := newTrade(s)
		if db.HasTable(tp) {
			db.Last(tp)
			log.Printf("已经从 %s 的表中获取了 UTC = %s， ID = %d\n", s, time.Unix(0, tp.UTC*1000000), tp.ID)
			heap.Push(&res, newRecord(s, tp.UTC, tp.ID))
		} else {
			db.CreateTable(tp)
			log.Printf("已经创建 %s 的表。\n", s)
			heap.Push(&res, newRecord(s, 0, 0))
		}
	}
	return &res
}

func (rs records) Len() int { return len(rs) }

func (rs records) Less(i, j int) bool {
	return rs[i].utc < rs[j].utc
}

func (rs records) Swap(i, j int) {
	rs[i], rs[j] = rs[j], rs[i]
}

// Push 往 rs 中放 record
func (rs *records) Push(x interface{}) {
	temp := x.(*record)
	*rs = append(*rs, temp)
}

// Pop 从 rs 中取出最优先的 record
func (rs *records) Pop() interface{} {
	temp := (*rs)[len(*rs)-1]
	*rs = (*rs)[0 : len(*rs)-1]
	return temp
}
