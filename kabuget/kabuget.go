package main

import (
	"database/sql"
	"flag"
	"log"
	"strconv"
	"time"

	"fmt"

	"local.packeges/com"
	_ "local.packeges/kabudata"

	_ "github.com/go-sql-driver/mysql"
)

func main() {

	wlogFileName := "kabuget_" + time.Now().Format("20060102_150405.log")
	com.LoggingSetting(wlogFileName)
	// logを実行する
	log.Println("処理開始")

	var (
		aDBUName  = flag.String("DBName", "kabu_test", "DB名")
		aDBUser   = flag.String("DBUser", "kabu", "DBユーザー")
		aDBPass   = flag.String("DBPass", "kabukabu", "DBパスワード")
		aDBServer = flag.String("DBServer", "localhost", "DBサーバ")
		aDBPort   = flag.String("DBPort ", "3306", "DBポート")
		aHonban   = flag.Bool("Honban", false, "True:本番環境")
		aPassword = flag.String("Password", "tkumi0312", "Kabuステーションのパスワード")
	)
	wDbString := *aDBUser + ":" + *aDBPass + "@tcp(" + *aDBServer + ":" + *aDBPort + ")/" + *aDBUName + "?parseTime=true&loc=Asia%2FTokyo"
	flag.Parse()
	log.Printf("DBStrring:" + wDbString)
	log.Println("Honban:" + strconv.FormatBool(*aHonban))
	log.Println("Password:" + *aPassword)

	//:db、2:db2、3:db3
	//SW := ""
	//0:本番、1:テスト
	//wDEBUG := 0

	db, err := sql.Open("mysql", wDbString)
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	wNow := time.Now()
	wTodayNum, wToday := com.GetDateNum(wNow)
	//wGetMaxDateNum := int(kabudata.GetMaxDateNum(db))

	log.Println("取得日付:" + wToday.Format("2006/01/02"))
	log.Println("取得日付NUM:" + strconv.Itoa(wTodayNum))
	//log.Println("取得済みNUM:" + strconv.Itoa(wGetMaxDateNum))

	var wSQL string

	//var wTkabukaBase kabudata.TkabukaBase

	var wCnt sql.NullInt64
	wSQL = "SELECT COUNT(*) FROM tkabuka WHERE HizukeNum = " + strconv.Itoa(wTodayNum)
	rowCnt, err := db.Query(wSQL)
	if err != nil {
		panic(err.Error())
	}
	if rowCnt.Next() {
		err = rowCnt.Scan(&wCnt)
		if err != nil {
			panic(err.Error())
		}
	}
	defer rowCnt.Close()

	if wCnt.Int64 > 0 {
		log.Println("株価は取得済みです。処理を中止します。")
		return
	}

	//株ステーションからトークンを取得
	wToken, err := com.GetToken(*aHonban, *aPassword)
	if err != nil {
		panic("Error")
	}
	fmt.Println(wToken)

	var wGetCnt int
	for i := 1001; i < 10000; i++ {
		wMeigaraCd := strconv.Itoa(i)
		wGetSymbol, err1 := com.GetSymbol(*aHonban, wToken, wMeigaraCd)
		if err1.Code > 0 {
			log.Println(wMeigaraCd + ":" + err1.Message)
		} else {
			log.Println(wGetSymbol)
			wGetBoard, err2 := com.GetBoard(*aHonban, wToken, wMeigaraCd)
			if err2.Code > 0 {
				log.Println(wMeigaraCd + ":" + err2.Message)
			} else {
				log.Println(wGetBoard)
				wGetCnt++
			}
		}
		time.Sleep(100 * time.Millisecond)
		if wGetCnt >= 40 {
			woUnregisterAll := com.GoUnregisterAll(*aHonban, wToken)
			log.Println(woUnregisterAll)
			time.Sleep(100 * time.Millisecond)
			wGetCnt = 0
		}
	}

}
