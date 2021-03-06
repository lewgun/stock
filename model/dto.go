package model

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/carusyte/stock/util"
	"log"
	"strings"
	"strconv"
	"github.com/pkg/errors"
	"math"
	"time"
	"gopkg.in/gorp.v2"
)

type DBTab string
type CYTP string

const (
	DAY   CYTP = "D"
	WEEK  CYTP = "W"
	MONTH CYTP = "M"
	M120  CYTP = "M120"
	M60   CYTP = "M60"
	M30   CYTP = "M30"
	M15   CYTP = "M15"
	M5    CYTP = "M5"
	M1    CYTP = "M1"
)

const (
	INDICATOR_DAY   DBTab = "indicator_d"
	INDICATOR_WEEK  DBTab = "indicator_w"
	INDICATOR_MONTH DBTab = "indicator_m"
	KLINE_DAY       DBTab = "kline_d"
	KLINE_DAY_NR    DBTab = "kline_d_n"
	KLINE_WEEK      DBTab = "kline_w"
	KLINE_MONTH     DBTab = "kline_m"
	KLINE_60M       DBTab = "kline_60m"
)

type Stock struct {
	Code             string
	Name             string
	Market           sql.NullString
	Industry         sql.NullString
	Area             sql.NullString
	Pe               sql.NullFloat64
	Pu               sql.NullFloat64
	Po               sql.NullFloat64
	Outstanding      sql.NullFloat64
	Totals           sql.NullFloat64
	TotalAssets      sql.NullFloat64
	LiquidAssets     sql.NullFloat64
	FixedAssets      sql.NullFloat64
	Reserved         sql.NullFloat64
	ReservedPerShare sql.NullFloat64
	Esp              sql.NullFloat64
	Bvps             sql.NullFloat64
	Pb               sql.NullFloat64
	TimeToMarket     sql.NullString
	Undp             sql.NullFloat64
	Perundp          sql.NullFloat64
	Rev              sql.NullFloat64
	Profit           sql.NullFloat64
	Gpr              sql.NullFloat64
	Npr              sql.NullFloat64
	Holders          sql.NullInt64
	Price            sql.NullFloat64
	Varate           sql.NullFloat64
	Var              sql.NullFloat64
	Xrate            sql.NullFloat64
	Volratio         sql.NullFloat64
	Ampl             sql.NullFloat64
	Turnover         sql.NullFloat64
	Accer            sql.NullFloat64
	CircMarVal       sql.NullFloat64
	UDate            sql.NullString
	UTime            sql.NullString
}

func (s *Stock) String() string {
	j, e := json.Marshal(s)
	if e != nil {
		fmt.Println(e)
	}
	return fmt.Sprintf("%v", string(j))
}

type Stocks struct {
	Map   map[string]*Stock
	List  []*Stock
	Codes []string
}

func (l *Stocks) Diff(a *Stocks) (same bool, diff []string) {
	if a == nil {
		return false, nil
	}
	if l.Size() == 0 && a.Size() == 0 {
		return true, []string{}
	}
	diff = make([]string, 0, int(math.Max(16, math.Abs(float64(l.Size()-a.Size())))))
	for _, c := range l.Codes {
		if _, exists := a.Map[c]; !exists {
			diff = append(diff, c)
		}
	}
	for _, c := range a.Codes {
		if _, exists := l.Map[c]; !exists {
			diff = append(diff, c)
		}
	}
	return len(diff) == 0, diff
}

func (l *Stocks) Size() int {
	return len(l.Codes)
}

func (l *Stocks) SetMarket(m string) {
	for _, s := range l.List {
		s.Market.Valid = true
		s.Market.String = m
	}
}

func (l *Stocks) Add(stks ... *Stock) {
	if stks == nil || len(stks) == 0 {
		return
	}
	if l.Codes == nil {
		l.Codes = make([]string, 0, 16)
	}
	if l.List == nil {
		l.List = make([]*Stock, 0, 16)
	}
	if l.Map == nil {
		l.Map = make(map[string]*Stock)
	}
	for _, s := range stks {
		l.Map[s.Code] = s
		l.List = append(l.List, s)
		l.Codes = append(l.Codes, s.Code)
	}
}

func (l *Stocks) String() string {
	j, e := json.Marshal(l)
	if e != nil {
		fmt.Println(e)
	}
	return fmt.Sprintf("%v", string(j))
}

func (l *Stocks) UnmarshalJSON(b []byte) error {
	var f interface{}
	json.Unmarshal(b, &f)
	m := f.(map[string]interface{})
	page := m["pageHelp"].(map[string]interface{})
	tot := int(page["total"].(float64))
	data := page["data"].([]interface{})
	if len(data) != tot {
		return fmt.Errorf("unmatched total numbers: %d/%d", len(data), tot)
	}
	l.List = make([]*Stock, len(data))
	l.Codes = make([]string, len(data))
	l.Map = make(map[string]*Stock, len(data))
	for i, da := range data {
		s := &Stock{}
		d := da.(map[string]interface{})
		if v, e := strconv.ParseFloat(d["totalFlowShares"].(string), 64); e == nil {
			s.Outstanding.Float64 = v / 10000.0
			s.Outstanding.Valid = true
		} else {
			return fmt.Errorf("failed to parse totalFlowShares: %+v, %+v", d["totalFlowShares"], e)
		}
		if v, ok := d["LISTING_DATE"].(string); ok {
			s.TimeToMarket.String = v
			s.TimeToMarket.Valid = true
		} else {
			return fmt.Errorf("failed to parse LISTING_DATE: %+v", d["LISTING_DATE"])
		}
		if v, ok := d["SECURITY_CODE_A"].(string); ok {
			s.Code = v
		} else {
			return fmt.Errorf("failed to parse SECURITY_CODE_A: %+v", d["SECURITY_CODE_A"])
		}
		if v, ok := d["SECURITY_ABBR_A"].(string); ok {
			s.Name = v
		} else {
			return fmt.Errorf("failed to parse SECURITY_ABBR_A: %+v", d["SECURITY_ABBR_A"])
		}
		if v, e := strconv.ParseFloat(d["totalShares"].(string), 64); e == nil {
			s.Totals.Float64 = v / 10000.0
			s.Totals.Valid = true
		} else {
			return fmt.Errorf("failed to parse totalShares: %+v, %+v", d["totalShares"], e)
		}
		dt, tm := util.TimeStr()
		s.UDate.Valid = true
		s.UTime.Valid = true
		s.UDate.String = dt
		s.UTime.String = tm
		l.List[i] = s
		l.Codes[i] = s.Code
		l.Map[s.Code] = s
	}
	return nil
}

type Xdxr struct {
	Code string
	Name string
	Idx  int
	//公告日期
	NoticeDate sql.NullString `db:"notice_date"`
	//报告期
	ReportYear sql.NullString `db:"report_year"`
	//董事会日期
	BoardDate sql.NullString `db:"board_date"`
	//每10股分红金额
	Divi sql.NullFloat64 `db:"divi"`
	//每10股分红金额（税后）
	DiviAtx sql.NullFloat64
	//分红截止日期
	DiviEndDate sql.NullString
	//分红率
	Dyr sql.NullFloat64 `db:"dyr"`
	//分红对象
	DiviTarget sql.NullString
	//每十股送红股
	SharesAllot sql.NullFloat64
	//红股上市日期
	SharesAllotDate sql.NullString
	//每十股转增股本
	SharesCvt sql.NullFloat64
	//转增股本上市日
	SharesCvtDate sql.NullString
	//派息股本基数
	SharesBase sql.NullInt64
	//股东大会日期
	GmsDate sql.NullString `db:"gms_date"`
	//实施日期
	ImplDate sql.NullString `db:"impl_date"`
	//分红方案说明
	Plan sql.NullString
	//股权登记日
	RegDate sql.NullString `db:"reg_date"`
	//除权除息日
	XdxrDate sql.NullString `db:"xdxr_date"`
	//股息到账日
	PayoutDate sql.NullString `db:"payout_date"`
	//最后交易日
	EndTrdDate sql.NullString
	//方案进度
	Progress sql.NullString `db:"progress"`
	//股利支付率 Dividend Payout Ratio
	Dpr sql.NullFloat64 `db:"dpr"`
	//股价刷新标记
	Xprice sql.NullString `db:"xprice"`
	//最后更新日期
	Udate sql.NullString
	//最后更新时间
	Utime sql.NullString
}

func (x *Xdxr) String() string {
	j, e := json.Marshal(x)
	if e != nil {
		fmt.Println(e)
	}
	return fmt.Sprintf("%v", string(j))
}

type Finance struct {
	Code string
	Year string
	//Earnings Per Share 每股收益
	Eps sql.NullFloat64
	//EPS Growth Rate Year-on-Year 每股收益同比增长率
	EpsYoy sql.NullFloat64 `db:"eps_yoy"`
	//Net Profit (1/10 Billion) 净利润（亿）
	Np sql.NullFloat64
	//Net Profit Growth Rate Year-on-Year 净利润同比增长率
	NpYoy sql.NullFloat64 `db:"np_yoy"`
	//Net Profit Ring Growth 净利润环比增长率
	NpRg sql.NullFloat64 `db:"np_rg"`
	//Net Profit After Deduction of Non-profits 扣除非经常性损益后的净利润
	NpAdn sql.NullFloat64 `db:"np_adn"`
	//Net Profit After Deduction of Non-profits Growth Rate Year-on-Year 扣非净利润同比增长率
	NpAdnYoy sql.NullFloat64 `db:"np_adn_yoy"`
	//Gross Revenue (1/10 Billion) 营业总收入（亿）
	Gr sql.NullFloat64
	//Gross Revenue Growth Rate Year-on-Year 营业总收入同比增长率
	GrYoy sql.NullFloat64 `db:"gr_yoy""`
	//Net Asset Value Per Share  每股净资产
	Navps sql.NullFloat64
	//Return on Equity 净资产收益率
	Roe sql.NullFloat64
	// ROE Growth Rate Year-on-Year 净资产收益率同比增长率
	RoeYoy sql.NullFloat64 `db:"roe_yoy"`
	//Return on Equity Diluted 净资产收益率-摊薄
	RoeDlt sql.NullFloat64 `db:"roe_dlt"`
	//Debt to Asset Ratio 资产负载比
	Dar sql.NullFloat64
	//Capital Reserves Per Share 每股资本公积
	Crps sql.NullFloat64
	//Undistributed Profit Per Share 每股未分配利润
	Udpps sql.NullFloat64
	// UDPPS Growth Rate Year-on-Year 每股未分配利润同比增长率
	UdppsYoy sql.NullFloat64 `db:"udpps_yoy"`
	//Operational Cash Flow Per Share 每股经营现金流
	Ocfps sql.NullFloat64
	// OCFPS Growth Rate Year-on-Year 每股经营现金流同比增长率
	OcfpsYoy sql.NullFloat64 `db:"ocfps_yoy"`
	//Gross Profit Margin 毛利率
	Gpm sql.NullFloat64
	//Net Profit Margin 净利率
	Npm sql.NullFloat64
	//Inventory Turnover Ratio 存货周转率
	Itr sql.NullFloat64
	//最后更新日期
	Udate sql.NullString
	//最后更新时间
	Utime sql.NullString
}

type FinReport struct {
	Items []*Finance
}

func (fin *FinReport) SetCode(code string) {
	for _, f := range fin.Items {
		f.Code = code
	}
}

func (fin *FinReport) UnmarshalJSON(b []byte) error {
	var f interface{}
	json.Unmarshal(b, &f)
	m := f.(map[string]interface{})
	titles := m["title"].([]interface{})
	iEps, iNp, iNpYoy, iNpRg, iNpAdn, iNpAdnYoy, iGr, iGrYoy, iNavps, iRoe, iRoeDlt, iAlr, iCrps, iUdpps, iOcfps,
	iGpm, iNpm, iItr := -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1
	mNp, mNpAdn, mGr := .1, .1, .1
	for i, t := range titles {
		v := fmt.Sprintf("%s", t)
		v = strings.Trim(v, "[]")
		v = strings.TrimSpace(v)
		switch v {
		case "基本每股收益 元":
			iEps = i
		case "净利润 万元":
			mNp = 0.0001
			fallthrough
		case "净利润 元":
			iNp = i
		case "净利润同比增长率 %":
			fallthrough
		case "净利润同比增长率":
			iNpYoy = i
		case "净利润环比增长率 %":
			fallthrough
		case "净利润环比增长率":
			iNpRg = i
		case "扣非净利润 万元":
			mNpAdn = 0.0001
			fallthrough
		case "扣非净利润 元":
			iNpAdn = i
		case "扣非净利润同比增长率 %":
			fallthrough
		case "扣非净利润同比增长率":
			iNpAdnYoy = i
		case "营业总收入 万元":
			mGr = 0.0001
			fallthrough
		case "营业总收入 元":
			iGr = i
		case "营业总收入同比增长率 %":
			fallthrough
		case "营业总收入同比增长率":
			iGrYoy = i
		case "每股净资产 元":
			iNavps = i
		case "净资产收益率 %":
			fallthrough
		case "净资产收益率":
			iRoe = i
		case "净资产收益率-摊薄 %":
			fallthrough
		case "净资产收益率-摊薄":
			iRoeDlt = i
		case "资产负债比率 %":
			fallthrough
		case "资产负债比率":
			iAlr = i
		case "每股资本公积金 元":
			iCrps = i
		case "每股未分配利润 元":
			iUdpps = i
		case "每股经营现金流 元":
			iOcfps = i
		case "销售毛利率 %":
			fallthrough
		case "销售毛利率":
			iGpm = i
		case "存货周转率":
			iItr = i
		case "销售净利率 %":
			fallthrough
		case "销售净利率":
			iNpm = i
		case `科目\时间`:
			//do nothing
		default:
			log.Printf("unidentified finance report item: %s", v)
		}
	}
	rpt := m["report"].([]interface{})
	for i, r := range rpt {
		if i == 0 {
			//parse year
			for _, iy := range r.([]interface{}) {
				fi := &Finance{}
				fi.Year = iy.(string)
				fin.Items = append(fin.Items, fi)
			}
		} else {
			//parse data
			for j, y := range r.([]interface{}) {
				if s, ok := y.(string); ok {
					fi := fin.Items[j]
					switch i {
					case iEps:
						fi.Eps = util.Str2Fnull(s)
					case iNp:
						fi.Np = util.Str2FBilMod(s, mNp)
					case iNpYoy:
						fi.NpYoy = util.Pct2Fnull(s)
					case iNpRg:
						fi.NpRg = util.Pct2Fnull(s)
					case iNpAdn:
						fi.NpAdn = util.Str2FBilMod(s, mNpAdn)
					case iNpAdnYoy:
						fi.NpAdnYoy = util.Pct2Fnull(s)
					case iGr:
						fi.Gr = util.Str2FBilMod(s, mGr)
					case iGrYoy:
						fi.GrYoy = util.Pct2Fnull(s)
					case iNavps:
						fi.Navps = util.Str2Fnull(s)
					case iRoe:
						fi.Roe = util.Pct2Fnull(s)
					case iRoeDlt:
						fi.RoeDlt = util.Pct2Fnull(s)
					case iAlr:
						fi.Dar = util.Pct2Fnull(s)
					case iCrps:
						fi.Crps = util.Str2Fnull(s)
					case iUdpps:
						fi.Udpps = util.Str2Fnull(s)
					case iOcfps:
						fi.Ocfps = util.Str2Fnull(s)
					case iGpm:
						fi.Gpm = util.Pct2Fnull(s)
					case iNpm:
						fi.Npm = util.Pct2Fnull(s)
					case iItr:
						fi.Itr = util.Str2Fnull(s)
					default:
						log.Printf("unidentified row index %d, %+v", i, y)
					}
				}
			}
		}
	}
	return nil
}

type Quote struct {
	Code   string `db:",size:6"`
	Date   string `db:",size:10"`
	Time   sql.NullString
	Klid   int
	Open   float64
	High   float64
	Close  float64
	Low    float64
	Volume sql.NullFloat64
	Amount sql.NullFloat64
	Xrate  sql.NullFloat64
	Varate sql.NullFloat64
	Ma5    sql.NullFloat64
	Ma10   sql.NullFloat64
	Ma20   sql.NullFloat64
	Ma30   sql.NullFloat64
	Udate  sql.NullString
	Utime  sql.NullString
}

func (q *Quote) String() string {
	j, e := json.Marshal(q)
	if e != nil {
		fmt.Println(e)
	}
	return fmt.Sprintf("%v", string(j))
}

type K60MinList struct {
	Quotes []*Quote
}

type Kline struct {
	Quote
	Factor sql.NullFloat64
}

type KlineW struct {
	Quote
}

type KlineM struct {
	Quote
}

type Indicator struct {
	Code  string `db:",size:6"`
	Date  string `db:",size:10"`
	Klid  int
	KDJ_K float64
	KDJ_D float64
	KDJ_J float64
	//最后更新日期
	Udate sql.NullString
	//最后更新时间
	Utime sql.NullString
}

type IndicatorW struct {
	Indicator
}

type IndicatorM struct {
	Indicator
}

func (k *K60MinList) UnmarshalJSON(b []byte) error {
	var f interface{}
	json.Unmarshal(b, &f)
	clist := f.(map[string]interface{})["chartlist"].([]interface{})
	k.Quotes = make([]*Quote, len(clist))
	for i, ci := range clist {
		im := ci.(map[string]interface{})
		q := new(Quote)
		k.Quotes[i] = q
		for k := range im {
			switch k {
			case "volume":
				q.Volume.Valid = true
				q.Volume.Float64 = im[k].(float64)
			case "open":
				q.Open = im[k].(float64)
			case "high":
				q.High = im[k].(float64)
			case "close":
				q.Close = im[k].(float64)
			case "low":
				q.Low = im[k].(float64)
			default:
				//do nothing
			}
		}
	}
	return nil
}

func (k *KlineW) String() string {
	j, e := json.Marshal(k)
	if e != nil {
		fmt.Println(e)
	}
	return fmt.Sprintf("%v", string(j))
}

type Klast struct {
	//Rt         string `json:"rt"`
	Num int `json:"num"`
	//Total      int `json:"total"`
	Start string                 `json:"start"`
	Year  map[string]interface{} `json:"year"`
	Name  string                 `json:"name"`
	Khist
	//IssuePrice float32 `json:"issuePrice"`
}

func (kl *Klast) UnmarshalJSON(b []byte) error {
	var f interface{}
	json.Unmarshal(b, &f)

	m := f.(map[string]interface{})

	for k := range m {
		switch k {
		case "num":
			kl.Num = int(m[k].(float64))
		case "start":
			kl.Start = m[k].(string)
		case "year":
			if y, ok := m[k].(map[string]interface{}); ok {
				kl.Year = y
			} else {
				kl.Year = make(map[string]interface{}, 0)
			}
		case "name:":
			kl.Name = m[k].(string)
		case "data":
			kl.Data = m[k].(string)
			//case "issuePrice":
			//	if v, ok := m[k].(string);ok{
			//		kl.IssuePrice =
			//	}
			//case "total":
			//case "rt":
		default:
			//do nothing
		}
	}
	return nil
}

type Khist struct {
	Data string `json:"data"`
}

type Ktoday struct {
	Quote
}

func (kt *Ktoday) UnmarshalJSON(b []byte) (e error) {
	defer func() {
		if r := recover(); r != nil {
			if er, ok := r.(error); ok {
				log.Println(er)
				e = errors.Wrap(er, fmt.Sprintf("failed to unmarshal Ktoday json: %s", string(b)))
			}
		}
	}()
	var f interface{}
	json.Unmarshal(b, &f)

	m := f.(map[string]interface{})

	for k := range m {
		qm := m[k].(map[string]interface{})
		if dt, ok := qm["1"].(string); ok {
			kt.Code = k[3:]
			kt.Date = dt[:4] + "-" + dt[4:6] + "-" + dt[6:]
			kt.Open = util.Str2F64(qm["7"].(string))
			kt.High = util.Str2F64(qm["8"].(string))
			kt.Low = util.Str2F64(qm["9"].(string))
			kt.Close = util.Str2F64(qm["11"].(string))
			kt.Volume = sql.NullFloat64{qm["13"].(float64), true}
			kt.Amount = sql.NullFloat64{util.Str2F64(qm["19"].(string)), true}
			kt.Xrate = sql.NullFloat64{util.Str2F64(qm["1968584"].(string)), true}
		} else {
			e = errors.Errorf("failed to parse Ktoday json: %s", string(b))
			return
		}
	}

	return nil
}

type IndcFeatRaw struct {
	Code    string
	Indc    string
	Cytp    string
	Bysl    string
	SmpDate string `db:"smp_date"`
	SmpNum  int    `db:"smp_num"`
	Fid     string
	Mark    float64
	Tspan   int
	Mpt     float64
	Remarks sql.NullString
	Udate   string
	Utime   string
}

func (indf *IndcFeatRaw) GenFid() string {
	indf.Fid = fmt.Sprintf("%s%s%s", indf.Cytp, indf.Bysl, strings.Replace(indf.SmpDate, "-", "", -1))
	return indf.Fid
}

func (indf *IndcFeatRaw) String() string {
	return fmt.Sprintf("|%s,%s,%s,%f,%d,%f|", indf.Code, indf.Fid, indf.Bysl, indf.Mark, indf.Tspan, indf.Mpt)
}

type IndcFeat struct {
	Indc    string
	Fid     string
	Cytp    string
	Bysl    string
	SmpNum  int `db:"smp_num"`
	FdNum   int `db:"fd_num"`
	Weight  float64
	Remarks sql.NullString
	Udate   string
	Utime   string
}

type KDJfd struct {
	Fid   string
	Seq   int
	K     float64
	D     float64
	J     float64
	Udate string
	Utime string
}

type KDJfdView struct {
	Indc, Fid, Bysl, Remarks string
	Cytp                     CYTP
	SmpNum, FdNum            int
	Weight                   float64
	K                        []float64
	D                        []float64
	J                        []float64
}

func (kfv *KDJfdView) Add(k, d, j float64) {
	kfv.K = append(kfv.K, k)
	kfv.D = append(kfv.D, d)
	kfv.J = append(kfv.J, j)
}

func (v *KDJfdView) String() string {
	j, e := json.Marshal(v)
	if e != nil {
		fmt.Println(e)
	}
	return fmt.Sprintf("%v", string(j))
}

type KDJfdRaw struct {
	Code  string
	Fid   string
	Klid  int
	K     float64
	D     float64
	J     float64
	Udate string
	Utime string
	Feat  *IndcFeatRaw
}

type KDJfdrView struct {
	Code    string
	SmpDate string
	SmpNum  int
	Klid    []int
	K       []float64
	D       []float64
	J       []float64
}

func (kfv *KDJfdrView) Add(klid int, k, d, j float64) {
	kfv.Klid = append(kfv.Klid, klid)
	kfv.K = append(kfv.K, k)
	kfv.D = append(kfv.D, d)
	kfv.J = append(kfv.J, j)
}

type KDJVStat struct {
	Code, Frmdt, Todt, Udate, Utime             string
	Dod, Sl, Sh, Bl, Bh, Sor, Bor, Smean, Bmean float64
	Scnt, Bcnt                                  int
}

type XQJson struct {
	Stock struct {
		Symbol string
	}
	Success string
	Chartlist []struct {
		Volume                                                                                int64
		Open, High, Low, Close, Chg, Percent, Turnrate, Ma5, Ma10, Ma20, Ma30, Dif, Dea, Macd float64
		LotVolume                                                                             int64 `json:"lot_volume"`
		Timestamp                                                                             int64
		Time                                                                                  string
	}
}

func (xqj *XQJson) Save(dbmap *gorp.DbMap, sklid int, table string) {
	if len(xqj.Chartlist) > 0 {
		valueStrings := make([]string, 0, len(xqj.Chartlist))
		valueArgs := make([]interface{}, 0, len(xqj.Chartlist)*13)
		var code string
		klid := sklid
		for _, q := range xqj.Chartlist {
			valueStrings = append(valueStrings, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, round(?,3), ?, ?)")
			valueArgs = append(valueArgs, xqj.Stock)
			valueArgs = append(valueArgs,
				time.Unix(q.Timestamp/int64(time.Microsecond), 0).Format("2006-01-02"))
			valueArgs = append(valueArgs, klid)
			valueArgs = append(valueArgs, q.Open)
			valueArgs = append(valueArgs, q.High)
			valueArgs = append(valueArgs, q.Close)
			valueArgs = append(valueArgs, q.Low)
			valueArgs = append(valueArgs, q.Volume)
			//valueArgs = append(valueArgs, q.Amount)
			//valueArgs = append(valueArgs, q.Xrate)
			//valueArgs = append(valueArgs, q.Varate)
			//valueArgs = append(valueArgs, q.Udate)
			//valueArgs = append(valueArgs, q.Utime)
			//code = q.Code
			klid++
		}
		stmt := fmt.Sprintf("INSERT INTO %s (code,date,klid,open,high,close,low,"+
			"volume,amount,xrate,varate,udate,utime) VALUES %s on duplicate key update date=values(date),"+
			"open=values(open),high=values(high),close=values(close),low=values(low),"+
			"volume=values(volume),amount=values(amount),xrate=values(xrate),varate=values(varate),udate=values"+
			"(udate),utime=values(utime)",
			table, strings.Join(valueStrings, ","))
		_, err := dbmap.Exec(stmt, valueArgs...)
		util.CheckErr(err, code+" failed to bulk insert "+table)
	}
}

// Set Code and Period before unmarshalling json data
type QQJson struct {
	Code, Period string
	Sklid        int
	Quotes       []*Quote
}

func (qj *QQJson) UnmarshalJSON(b []byte) error {
	var (
		f      interface{}
		m      map[string]interface{}
		retcde float64
		msg    string
		ok     bool
		e      error
	)
	e = json.Unmarshal(b, &f)
	if e != nil {
		return errors.Wrapf(e, "%s %s failed to unmarshal json data", qj.Code, qj.Period)
	}
	if m, ok = f.(map[string]interface{}); !ok {
		return errors.Errorf("unrecognized data structure: %+v", f)
	}
	retcde = m["code"].(float64)
	msg = m["msg"].(string)
	if retcde != 0 {
		return errors.Errorf("server failed with code %d, msg: %s", retcde, msg)
	}
	if cdat, exists := m["data"].(map[string]interface{})[qj.Code]; !exists {
		return errors.Errorf("unrecognized data structure: %+v", f)
	} else {
		if pdat, exists := cdat.(map[string]interface{})[qj.Period]; !exists {
			return errors.Errorf("unrecognized data structure: %+v", f)
		} else {
			ps := pdat.([]interface{})
			qj.Quotes = make([]*Quote, len(ps))
			klid := qj.Sklid
			dt, tm := util.TimeStr()
			for i, pd := range ps {
				pa := pd.([]interface{})
				q := new(Quote)
				q.Code = qj.Code
				q.Date = pa[0].(string)
				q.Klid = klid
				q.Open, e = strconv.ParseFloat(pa[1].(string), 64)
				if e != nil {
					return errors.Wrapf(e, "failed to parse OPEN value at index %d", i)
				}
				q.Close, e = strconv.ParseFloat(pa[2].(string), 64)
				if e != nil {
					return errors.Wrapf(e, "failed to parse CLOSE value at index %d", i)
				}
				q.High, e = strconv.ParseFloat(pa[3].(string), 64)
				if e != nil {
					return errors.Wrapf(e, "failed to parse HIGH value at index %d", i)
				}
				q.Low, e = strconv.ParseFloat(pa[4].(string), 64)
				if e != nil {
					return errors.Wrapf(e, "failed to parse LOW value at index %d", i)
				}
				q.Volume.Valid = true
				q.Volume.Float64, e = strconv.ParseFloat(pa[2].(string), 64)
				if e != nil {
					return errors.Wrapf(e, "failed to parse Volume value at index %d", i)
				}
				q.Udate.Valid = true
				q.Utime.Valid = true
				q.Udate.String = dt
				q.Utime.String = tm
				klid++
				qj.Quotes[i] = q
			}
		}
	}
	return nil
}

func (qj *QQJson) Save(dbmap *gorp.DbMap, table string) {
	if len(qj.Quotes) > 0 {
		valueStrings := make([]string, 0, len(qj.Quotes))
		valueArgs := make([]interface{}, 0, len(qj.Quotes)*10)
		for _, q := range qj.Quotes {
			valueStrings = append(valueStrings, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
			valueArgs = append(valueArgs, q.Code)
			valueArgs = append(valueArgs, q.Date)
			valueArgs = append(valueArgs, q.Klid)
			valueArgs = append(valueArgs, q.Open)
			valueArgs = append(valueArgs, q.High)
			valueArgs = append(valueArgs, q.Close)
			valueArgs = append(valueArgs, q.Low)
			valueArgs = append(valueArgs, q.Volume)
			valueArgs = append(valueArgs, q.Udate)
			valueArgs = append(valueArgs, q.Utime)
		}
		stmt := fmt.Sprintf("INSERT INTO %s (code,date,klid,open,high,close,low,"+
			"volume,udate,utime) VALUES %s on duplicate key update date=values(date),"+
			"open=values(open),high=values(high),close=values(close),low=values(low),"+
			"volume=values(volume),udate=values(udate),utime=values(utime)",
			table, strings.Join(valueStrings, ","))
		_, err := dbmap.Exec(stmt, valueArgs...)
		util.CheckErr(err, qj.Code+" failed to bulk insert "+table)
	}
}

// Index List
type IdxLst struct {
	Code, Name, Src string
}
