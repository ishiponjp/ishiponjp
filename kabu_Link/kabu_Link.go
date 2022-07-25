package main

import (
	"database/sql"
	"flag"
	"fmt"
	"strings"
	"strconv"

	"log"
	"time"

	"runtime"
	"sync"

	_ "github.com/go-sql-driver/mysql"
	"local.packeges/com"
)

const KomokuGen1 = "MeigaraCd,Hizuke,Meisyo,Sijyo,Taisyaku,HizukeNum,HazimariNe,YasuNe,TakaNe,OwariNe,Dekidaka,Sagaku,SagakuRitu,ZenOwariNe,Jika,Tani,MinKabuka,MaxKabuka,BaibaiSougaku"	
const KomokuGen2 = "MeigaraCd,Hizuke,Meisyo,Sijyo,Sinyo,HizukeNum,HazimariNe,YasuNe,TakaNe,OwariNe,Dekidaka,Sagaku,SagakuRitu,ZenOwariNe,Jika,Tani,MinKabuka,MaxKabuka,BaibaiSougaku"	


//MeigaraCnt is struct
type MeigaraCnt struct {
	MeigaraCd sql.NullString
	Cnt       sql.NullInt64
}

//MeigaraCnt is struct
type Tkabuka struct {
	MeigaraCd 	sql.NullString
	Hizuke 		sql.NullString
	Meisyo 		sql.NullString
	Sijyo 		sql.NullString
	Sinyo		sql.NullInt64
	HizukeNum	sql.NullInt64
	HazimariNe	sql.NullFloat64
	YasuNe		sql.NullFloat64
	TakaNe		sql.NullFloat64
	OwariNe		sql.NullFloat64
	Dekidaka	sql.NullFloat64
	Sagaku		sql.NullFloat64
	SagakuRitu	sql.NullFloat64
	ZenOwariNe	sql.NullFloat64
	Jika		sql.NullFloat64
	Tani		sql.NullFloat64
	MinKabuka	sql.NullFloat64
	MaxKabuka	sql.NullFloat64
	BaibaiSougaku	sql.NullFloat64
}


func scanTkabuka(iRow *sql.Rows)(oTkabuka Tkabuka){

	var wTkabuka1 Tkabuka
	err := iRow.Scan(&wTkabuka1.MeigaraCd,
		&wTkabuka1.Hizuke,
		&wTkabuka1.Meisyo,
		&wTkabuka1.Sijyo,
		&wTkabuka1.Sinyo,
		&wTkabuka1.HizukeNum,
		&wTkabuka1.HazimariNe,
		&wTkabuka1.YasuNe,
		&wTkabuka1.TakaNe,
		&wTkabuka1.OwariNe,
		&wTkabuka1.Dekidaka,
		&wTkabuka1.Sagaku,
		&wTkabuka1.SagakuRitu,
		&wTkabuka1.ZenOwariNe,
		&wTkabuka1.Jika,
		&wTkabuka1.Tani,
		&wTkabuka1.MinKabuka,
		&wTkabuka1.MaxKabuka,
		&wTkabuka1.BaibaiSougaku,)
	if err != nil {
		panic(err)
	}
	oTkabuka = wTkabuka1

	return
}

func convTkabuka(iTkabuka Tkabuka,iFrom int,iTo int)(oTkabuka Tkabuka){

	//1：制度信用（6ヶ月）、2：一般信用（無制限）、3：一般信用（14日）、
	//4：一般信用（いちにち）
	

	oTkabuka = iTkabuka

	if iFrom == iTo {
		return
	}

	if iFrom == 1 && iTo == 2 {
		oTkabuka.Jika.Float64 = iTkabuka.Jika.Float64 / 1000000
		oTkabuka.BaibaiSougaku.Float64 = iTkabuka.BaibaiSougaku.Float64 / 1000
	} else if iFrom == 2 && iTo == 1 {
		oTkabuka.Jika.Float64 = iTkabuka.Jika.Float64 * 1000000
		oTkabuka.BaibaiSougaku.Float64 = iTkabuka.BaibaiSougaku.Float64 * 1000
	}

	return
}

func getInsertSql(iTkabuka Tkabuka,iKoumoku string,iFrom int,iTo int)(oSql string){

	wKabuka := convTkabuka(iTkabuka,iFrom,iTo)

	var wVal []string
	wVal = append(wVal, com.GetSqlValueString(wKabuka.MeigaraCd))
	wVal = append(wVal, com.GetSqlValueString(wKabuka.Hizuke))
	wVal = append(wVal, com.GetSqlValueString(wKabuka.Meisyo))
	wVal = append(wVal, com.GetSqlValueString(wKabuka.Sijyo))
	wVal = append(wVal, com.GetSqlValueInt64(wKabuka.Sinyo))
	wVal = append(wVal, com.GetSqlValueInt64(wKabuka.HizukeNum))
	wVal = append(wVal, com.GetSqlValueFloat64(wKabuka.HazimariNe))
	wVal = append(wVal, com.GetSqlValueFloat64(wKabuka.YasuNe))
	wVal = append(wVal, com.GetSqlValueFloat64(wKabuka.TakaNe))
	wVal = append(wVal, com.GetSqlValueFloat64(wKabuka.OwariNe))
	wVal = append(wVal, com.GetSqlValueFloat64(wKabuka.Dekidaka))
	wVal = append(wVal, com.GetSqlValueFloat64(wKabuka.Sagaku))
	wVal = append(wVal, com.GetSqlValueFloat64(wKabuka.SagakuRitu))
	wVal = append(wVal, com.GetSqlValueFloat64(wKabuka.ZenOwariNe))
	wVal = append(wVal, com.GetSqlValueFloat64(wKabuka.Jika))
	wVal = append(wVal, com.GetSqlValueFloat64(wKabuka.Tani))
	wVal = append(wVal, com.GetSqlValueFloat64(wKabuka.MinKabuka))
	wVal = append(wVal, com.GetSqlValueFloat64(wKabuka.MaxKabuka))
	wVal = append(wVal, com.GetSqlValueFloat64(wKabuka.BaibaiSougaku))
	wSQLValue := strings.Join(wVal[:], ",")

	oSql = "INSERT INTO tkabuka (" + iKoumoku + ") VALUES (" + wSQLValue + ");"

	return oSql
}


func matchMeigara(iDb1 *sql.DB,iDb2 *sql.DB,iMeigaraCd string,
	iKoumokuGen1 int,iKoumokuGen2 int){

	log.Println("matchMeigara:" + iMeigaraCd)

	var wKomoku1 string
	var wKomoku2 string
	if iKoumokuGen1 == 1 {
		wKomoku1 = KomokuGen1
	} else if iKoumokuGen1 == 2 {
		wKomoku1 = KomokuGen2
	}
	if iKoumokuGen2 == 1 {
		wKomoku2 = KomokuGen1
	} else if iKoumokuGen2 == 2 {
		wKomoku2 = KomokuGen2
	}
	//wKomoku1 := "MeigaraCd,Hizuke,Meisyo,Sijyo,Sinyo,HizukeNum,HazimariNe,YasuNe,TakaNe,OwariNe,Dekidaka,Sagaku,SagakuRitu,ZenOwariNe,Jika,Tani,MinKabuka,MaxKabuka,BaibaiSougaku"	
	//wKomoku2 := "MeigaraCd,Hizuke,Meisyo,Sijyo,Taisyaku,HizukeNum,HazimariNe,YasuNe,TakaNe,OwariNe,Dekidaka,Sagaku,SagakuRitu,ZenOwariNe,Jika,Tani,MinKabuka,MaxKabuka,BaibaiSougaku"	
	wSQL1 := fmt.Sprintf("SELECT " + wKomoku1 + " FROM tkabuka WHERE MeigaraCd = '%s' ORDER BY Hizuke", iMeigaraCd)
	wSQL2 := fmt.Sprintf("SELECT " + wKomoku2 + " FROM tkabuka WHERE MeigaraCd = '%s' ORDER BY Hizuke", iMeigaraCd)

	row1, err := iDb1.Query(wSQL1)
	if err != nil {
		panic(err.Error())
	}
	defer row1.Close()

	var wTkabukaPool1 []Tkabuka

	// Scan
	for row1.Next() {
		wTkabuka := scanTkabuka(row1)
		wTkabukaPool1 = append(wTkabukaPool1, wTkabuka)
	}

	row2, err2 := iDb2.Query(wSQL2)
	if err2 != nil {
		panic(err2.Error())
	}
	defer row2.Close()

	var wTkabukaPool2 []Tkabuka

	// Scan
	for row2.Next() {
		wTkabuka := scanTkabuka(row2)
		wTkabukaPool2 = append(wTkabukaPool2, wTkabuka)
	}	

	//マッチング処理
	var i1 int
	var i2 int
	//DB１のあまり分⇒DB2へinsertする
	var InsertSql1 []string
	//DB２のあまり分⇒DB1へinsertする
	var InsertSql2 []string

	var wDateNum1 int
	var wDateNum2 int


	for i1 < len(wTkabukaPool1) || i2 < len(wTkabukaPool2){
		var wSql string
		if i1 >= len(wTkabukaPool1) {
			wDateNum1 = 99999999
		} else {
			wHizuke := wTkabukaPool1[i1].Hizuke.String
			wDateNum1,_ = strconv.Atoi(wHizuke[0:4] + wHizuke[5:7] + wHizuke[8:10])
		}
		if i2 >= len(wTkabukaPool2) {
			wDateNum2 = 99999999
		} else {
			wHizuke := wTkabukaPool2[i2].Hizuke.String
			wDateNum2,_ = strconv.Atoi(wHizuke[0:4] + wHizuke[5:7] + wHizuke[8:10])
		}

		if wDateNum1 < wDateNum2 {
			wSql = getInsertSql(wTkabukaPool1[i1],wKomoku2,iKoumokuGen1,iKoumokuGen2)
			InsertSql1 = append(InsertSql1, wSql)
			i1++
		} else if  wDateNum1 > wDateNum2 {
			wSql = getInsertSql(wTkabukaPool2[i2],wKomoku1,iKoumokuGen2,iKoumokuGen1)
			InsertSql2 = append(InsertSql2, wSql)
			i2++
		} else {
			i1++
			i2++
		}
	}

	for _, wSql := range InsertSql1 {
		_, err := iDb2.Exec(wSql)
		if err != nil {
			panic(err.Error())
		}
	}
	log.Println("Insert1:", iMeigaraCd, len(InsertSql1))

	for _, wSql := range InsertSql2 {
		_, err := iDb1.Exec(wSql)
		if err != nil {
			panic(err.Error())
		}
	}
	log.Println("Insert2:", iMeigaraCd, len(InsertSql2))	

	return
}


func main() {

	wlogFileName := "kabuLink_" + time.Now().Format("20060102_150405.log")
	com.LoggingSetting(wlogFileName)
	// logを実行する
    log.Println("処理開始")

	var (
		aDBStrring1 = flag.String("DBStrring1", "kabu:kabukabu@tcp(localhost:3306)/kabu_test", "DB接続文字列")
		aDBGen1 = flag.Int("DBGen1", 2, "DB世代,1:abee-t20/2:Ideapad530")
		aDBStrring2 = flag.String("DBStrring2", "kabu:kabukabu@tcp(abee-t20:3306)/kabu_test", "DB接続文字列")
		aDBGen2 = flag.Int("DBGen2", 1, "DB世代,1:abee-t20/2:Ideapad530")
		aDebug     = flag.Bool("Debug", false, "True:デバッグモード")
	)
	flag.Parse()
	fmt.Printf("DBStrring1: %v\n", *aDBStrring1)
	fmt.Printf("DBGen1: %v\n", *aDBGen1)
	fmt.Printf("DBStrring2: %v\n", *aDBStrring2)
	fmt.Printf("DBGen2: %v\n", *aDBGen2)
	fmt.Printf("Debug: %v\n", *aDebug)

	if *aDBGen1 != 1 && *aDBGen1 != 2 {
		log.Println("DBGen1は１か２を指定してください",*aDBGen1)
	}

	if *aDBGen2 != 1 && *aDBGen2 != 2 {
		log.Println("DBGen1は１か２を指定してください",*aDBGen2)
	}

	//:db、2:db2、3:db3
	//SW := ""
	//0:本番、1:テスト
	//wDEBUG := 0

	var wSQL string
	var wkabukaCntPool1 []MeigaraCnt
	var wkabukaCntPool2 []MeigaraCnt

	db1, err1 := sql.Open("mysql", *aDBStrring1)
	if err1 != nil {
		panic(err1.Error())
	}
	defer db1.Close()
	//日付番号セット
	wSQL = "SELECT MeigaraCd,Count(*) FROM tkabuka GROUP BY MeigaraCd ORDER BY MeigaraCd"

	rowSums1, err11 := db1.Query(wSQL)
	if err11 != nil {
		panic(err11.Error())
	}

	for rowSums1.Next() {
		var	wMeigaraCnt MeigaraCnt

		err12 := rowSums1.Scan(&wMeigaraCnt.MeigaraCd,&wMeigaraCnt.Cnt)
		if err12 != nil {
			panic(err12.Error())
		}
		wkabukaCntPool1 = append(wkabukaCntPool1, wMeigaraCnt)
	}
	rowSums1.Close()

	db2, err2 := sql.Open("mysql", *aDBStrring2)
	if err2 != nil {
		panic(err2.Error())
	}
	defer db2.Close()

	rowSums2, err2 := db2.Query(wSQL)
	if err2 != nil {
		panic(err2.Error())
	}

	for rowSums2.Next() {
		var	wMeigaraCnt MeigaraCnt

		err2 := rowSums2.Scan(&wMeigaraCnt.MeigaraCd,&wMeigaraCnt.Cnt)
		if err2 != nil {
			panic(err2.Error())
		}
		wkabukaCntPool2 = append(wkabukaCntPool2, wMeigaraCnt)
	}
	rowSums2.Close()

	//マッチング処理
	var i1 int
	var i2 int
	var MeigaraCdPool []string


	for i1 < len(wkabukaCntPool1) && i2 < len(wkabukaCntPool2){
		if i1 >= len(wkabukaCntPool1) {
			//i1が最後まで到達
			MeigaraCdPool = append(MeigaraCdPool, wkabukaCntPool2[i2].MeigaraCd.String)
			i2++
		} else if i2 >= len(wkabukaCntPool2) {
			//i2が最後まで到達
			MeigaraCdPool = append(MeigaraCdPool, wkabukaCntPool1[i1].MeigaraCd.String)
			i1++
		} else if wkabukaCntPool1[i1].MeigaraCd.String == wkabukaCntPool2[i2].MeigaraCd.String {
			if wkabukaCntPool1[i1].Cnt.Int64 != wkabukaCntPool2[i2].Cnt.Int64 {
				MeigaraCdPool = append(MeigaraCdPool, wkabukaCntPool1[i1].MeigaraCd.String )
			}
			i1++
			i2++
		} else if wkabukaCntPool1[i1].MeigaraCd.String < wkabukaCntPool2[i2].MeigaraCd.String {
			//i1が小さい
			MeigaraCdPool = append(MeigaraCdPool, wkabukaCntPool1[i1].MeigaraCd.String)
			i1++
		} else {
			MeigaraCdPool = append(MeigaraCdPool, wkabukaCntPool2[i2].MeigaraCd.String)
			i2++
		}

	}

	var wg sync.WaitGroup
	var cpus int
	if *aDebug {
		cpus = 1
	} else {
		cpus = int(float64(runtime.NumCPU()) * 0.75)
	}
	semaphore := make(chan string, cpus)

	for _,wMeigaraCd := range MeigaraCdPool{

		wg.Add(1)
		go func(wMeigaraCd string) {
			defer wg.Done()
			semaphore <- wMeigaraCd
			matchMeigara(db1,db2,wMeigaraCd,*aDBGen1,*aDBGen2)
			<-semaphore
		}(wMeigaraCd)
	}
	wg.Wait()

	return
}
