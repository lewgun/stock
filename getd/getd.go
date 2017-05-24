package getd

import "time"

func Get(){
	start := time.Now()
	defer stop("GETD_TOTAL", start)
	stks := GetStockInfo()
	stop("STOCK_LIST", start)

	stgx := time.Now()
	GetXDXRs(stks)
	stop("GET_XDXR", stgx)

	stgfi := time.Now()
	GetFinance(stks)
	stop("GET_FINANCE", stgfi)

	stgkl := time.Now()
	GetKlines(stks)
	stop("GET_KLINES", stgkl)

	stci := time.Now()
	CalcIndics(stks)
	stop("CALC_INDICS", stci)
}

func stop(code string, start time.Time) {
	ss := start.Format("2006-01-02 15:04:05")
	end := time.Now().Format("2006-01-02 15:04:05")
	dur := time.Since(start).Seconds()
	log.Printf("%s Complete. Time Elapsed: %f sec", code, dur)
	dbmap.Exec("insert into stats (code, start, end, dur) values (?, ?, ?, ?) "+
		"on duplicate key update start=values(start), end=values(end), dur=values(dur)",
		code, ss, end, dur)
}
