package main

import (
	"database/sql"
	"flag"
	"fmt"
	"math"
	"sort"

	"runtime"
	"sync"

	"local.packeges/com"

	_ "github.com/go-sql-driver/mysql"
)



// 構造体のスライス
type VkabukaAvePool []com.VkabukaAve
// 以下インタフェースを満たす

func (p VkabukaAvePool) Len() int {
	return len(p)
}

func (p VkabukaAvePool) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p VkabukaAvePool) Less(i, j int) bool {
	if p[i].HizukeInt < p[j].HizukeInt {
		return true
	} else if p[i].HizukeInt == p[j].HizukeInt {
		if p[i].MeigaraCdInt < p[j].MeigaraCdInt {
			return true
		} else {
			return false
		}
	} else {
		return false
	}
}

func hanNanpin(iKoza float64,iTiming string,iJyouken com.TJyoukenGen,iBaibaiKekka com.TsimbaibaikekkaGen,iKabuka com.VkabukaAve)(oKoza float64,oBaibaiKekka com.TsimbaibaikekkaGen){

	oKoza = iKoza
	oBaibaiKekka =iBaibaiKekka

	if oBaibaiKekka.UriIdx > 0 {
		//売られている場合、終わる
		return
	}

	if iJyouken.Nanpin == 1 {
		//ナンピンが１００％の場合、ナンピン無し
		return
	}

	if iBaibaiKekka.KaiIdxSl[iJyouken.NanpinLvl] > 0 {
		//ナンピンが最後までされた場合、終わる
		return
	}

	wKaiKingaku := iBaibaiKekka.KaiKabuka * iJyouken.Nanpin

	var wKabuka float64
	if iTiming == "O" {
		//終値でナンピン
		wKabuka = iKabuka.OwariNe
	} else if iTiming == "Y" {
		wKabuka = iKabuka.HazimariNeF1
	} else {
		//タイミング誤り
		return
	}

	if iJyouken.KBN == "K" {
		if wKaiKingaku < wKabuka {
			//買い株価以上の場合、終わる
			return
		}
	} else if iJyouken.KBN == "U" {
		if wKaiKingaku > wKabuka {
			//売り株価未満、終わる
			return
		}
	} else {
		//区分指定あやまり
		return
	}

	i := 1
	for (oBaibaiKekka.KaiIdxSl[i] > 0 ){
		i++
	}

	//買い増し
	oBaibaiKekka.KaiIdxSl[i] = iKabuka.HizukeNum
	oBaibaiKekka.KaiHizukeSl[i] = iKabuka.Hizuke
	oBaibaiKekka.KaiKabukaSl[i] = wKabuka
	oBaibaiKekka.KaiKabusuSl[i] = oBaibaiKekka.KaiKabusu
	oBaibaiKekka.KaiKabuka = (oBaibaiKekka.KaiKabuka + oBaibaiKekka.KaiKabukaSl[i]) / 2
	oBaibaiKekka.KaiKabugakuSl[i] = oBaibaiKekka.KaiKabukaSl[i] * float64(oBaibaiKekka.KaiKabusuSl[i])
	oBaibaiKekka.KaiTimingSl[i] = iTiming
	oBaibaiKekka.KaiKabugaku += oBaibaiKekka.KaiKabugakuSl[i]
	oBaibaiKekka.KaiKabusu += oBaibaiKekka.KaiKabusuSl[i]
	oBaibaiKekka.NanpinYotei = oBaibaiKekka.KaiKabuka * iJyouken.Nanpin
	oBaibaiKekka.UriYotei = oBaibaiKekka.KaiKabuka * iJyouken.UriRitu
	if iJyouken.UriRitu != iJyouken.UriSasiNeRitu {
		oBaibaiKekka.UriSasiYotei = oBaibaiKekka.KaiKabuka * iJyouken.UriSasiNeRitu
	}

	oBaibaiKekka.TesuRyo += com.GetTesuRyo(iJyouken.KBN, oBaibaiKekka.KaiKabugakuSl[i])

	//買い増しがあった。最大と最小をリセットする。
	oBaibaiKekka.MinKabuka = iKabuka.OwariNe
	oBaibaiKekka.MaxKabuka = iKabuka.OwariNe

	wKonyu := oBaibaiKekka.KaiKabukaSl[i] * float64(oBaibaiKekka.KaiKabusuSl[i])
	oKoza -= wKonyu
	return
}


func hanSel(iKoza float64,iTiming string,iJyouken com.TJyoukenGen,iBaibaiKekka com.TsimbaibaikekkaGen,iKabuka com.VkabukaAve)(oKoza float64,oBaibaiKekka com.TsimbaibaikekkaGen){

	oKoza = iKoza
	oBaibaiKekka =iBaibaiKekka

	if oBaibaiKekka.UriIdx > 0 {
		//売られている場合、終わる
		return
	}

	var wUriNe float64
	if iTiming == "O" {
		if iJyouken.KBN == "K" && oBaibaiKekka.UriYotei <= iKabuka.OwariNe {
			wUriNe = iKabuka.OwariNe
		} else if iJyouken.KBN == "U" && oBaibaiKekka.UriYotei >= iKabuka.OwariNe   {
			wUriNe = iKabuka.OwariNe
		} else {
			return
		}
	} else if iTiming == "Y"  {
		if iJyouken.KBN == "K" && oBaibaiKekka.UriYotei <= iKabuka.HazimariNeF1 {
			wUriNe = iKabuka.HazimariNeF1
		} else if iJyouken.KBN == "U" && oBaibaiKekka.UriYotei >= iKabuka.HazimariNeF1   {
			wUriNe = iKabuka.HazimariNeF1
		} else {
			return
		}
	} else if iTiming == "S"  {
		if oBaibaiKekka.UriSasiYotei  == 0 {
			return
		}
		if iKabuka.YasuNeF1 <= oBaibaiKekka.UriSasiYotei && iKabuka.TakaNeF1 >= oBaibaiKekka.UriSasiYotei {
			wUriNe = oBaibaiKekka.UriSasiYotei
		} else if iJyouken.KBN == "K" && iKabuka.YasuNeF1 >= oBaibaiKekka.UriSasiYotei {
			wUriNe = iKabuka.YasuNeF1
		} else if iJyouken.KBN == "U" && iKabuka.TakaNeF1 <= oBaibaiKekka.UriSasiYotei {
			wUriNe = iKabuka.TakaNeF1
		} else {
			return
		}
	}

	oBaibaiKekka.UriIdx = iKabuka.HizukeNum
	oBaibaiKekka.UriHizuke = iKabuka.Hizuke
	oBaibaiKekka.DayAll = oBaibaiKekka.UriIdx - oBaibaiKekka.KaiIdxSl[0]
	oBaibaiKekka.UriKabuka = wUriNe
	oBaibaiKekka.UriKabugaku = oBaibaiKekka.UriKabuka * float64(oBaibaiKekka.KaiKabusu)
	oBaibaiKekka.UriTiming = iTiming

	var wDay int
	wDay = oBaibaiKekka.UriIdx - oBaibaiKekka.KaiIdxSl[0]
	oBaibaiKekka.Kinri = com.GetKinri(iJyouken.KBN, oBaibaiKekka.KaiKabugakuSl[0], wDay)
	if oBaibaiKekka.KaiIdxSl[1] > 0 {
		wDay = oBaibaiKekka.UriIdx - oBaibaiKekka.KaiIdxSl[1]
		oBaibaiKekka.Kinri += com.GetKinri(iJyouken.KBN, oBaibaiKekka.KaiKabugakuSl[1], wDay)
	}
	if oBaibaiKekka.KaiIdxSl[2] > 0 {
		wDay = oBaibaiKekka.UriIdx - oBaibaiKekka.KaiIdxSl[2]
		oBaibaiKekka.Kinri += com.GetKinri(iJyouken.KBN, oBaibaiKekka.KaiKabugakuSl[2], wDay)
	}
	if oBaibaiKekka.KaiIdxSl[3] > 0 {
		wDay = oBaibaiKekka.UriIdx - oBaibaiKekka.KaiIdxSl[3]
		oBaibaiKekka.Kinri += com.GetKinri(iJyouken.KBN, oBaibaiKekka.KaiKabugakuSl[3], wDay)
	}

	oBaibaiKekka.TesuRyo += com.GetTesuRyo(iJyouken.KBN, oBaibaiKekka.UriKabugaku)

	if oBaibaiKekka.KaiKabugaku > 0 {
		if iJyouken.KBN == "K" {
			oBaibaiKekka.Soneki = oBaibaiKekka.UriKabugaku - oBaibaiKekka.KaiKabugaku - oBaibaiKekka.TesuRyo - oBaibaiKekka.Kinri
		} else {
			oBaibaiKekka.Soneki = oBaibaiKekka.KaiKabugaku - oBaibaiKekka.UriKabugaku - oBaibaiKekka.TesuRyo - oBaibaiKekka.Kinri
		}
		oBaibaiKekka.SonekiRitu = (oBaibaiKekka.Soneki/oBaibaiKekka.KaiKabugaku ) * 100
	}

	if iJyouken.KBN == "K" {
		oKoza += (oBaibaiKekka.UriKabugaku + oBaibaiKekka.NaninKoza)
	} else {
		oKoza += (oBaibaiKekka.KaiKabugaku + oBaibaiKekka.KaiKabugaku - oBaibaiKekka.UriKabugaku + oBaibaiKekka.NaninKoza)
	}

	return

}

func runBuySim(iSW string, iJyouken com.TJyoukenGen, iKabukas []com.VkabukaAve) (com.TJyoukenGen, []com.TsimbaibaikekkaGen, []com.TJyokenLogGen) {

	var wSimbaibaikekkaList []com.TsimbaibaikekkaGen
	var wJyoukenLogs []com.TJyokenLogGen

	wHizuke := ""
	wKoza := iJyouken.StaKoza

	//wKozaOff := false        //true:口座無視、false：口座残高確認
	//wNanpinKozaOff := false  //true:ナンピンを信用で買う false:現金
	//wNanpinKozaZanOn := true //True:現金でナンピン用のお金を残す

	wHoyu := make(map[string]com.TsimbaibaikekkaGen)
	wCnt := 0
	var wBuy bool
	wUriNe := 0.0
	for _, wKabuka := range iKabukas {

		if wHizuke != wKabuka.Hizuke {
			if len(wHizuke) > 0 {
				var wJyoukenLog com.TJyokenLogGen
				wJyoukenLog.GenHizuke = iJyouken.GenHizuke
				wJyoukenLog.GenHizukeNum = iJyouken.GenHizukeNum
				wJyoukenLog.No = iJyouken.ScenarioNo
				wJyoukenLog.Koza = wKoza
				wJyoukenLog.Hizuke = wHizuke
				for _, wJyouken := range wHoyu {
					wJyoukenLog.KaiGaku += wJyouken.KaiKabugaku
					wJyoukenLog.HyokaGaku += wJyouken.UriKabugaku
				}
				if wJyoukenLog.KaiGaku > 0 {
					if iJyouken.KBN == "K" {
						wJyoukenLog.Soneki = wJyoukenLog.HyokaGaku - wJyoukenLog.KaiGaku
					} else {
						wJyoukenLog.Soneki = wJyoukenLog.KaiGaku - wJyoukenLog.HyokaGaku
					}
					wJyoukenLog.SonekiRitu = (wJyoukenLog.Soneki/wJyoukenLog.KaiGaku) * 100
				}
				wJyoukenLog.KaiGakuGoukei = wJyoukenLog.Koza + wJyoukenLog.KaiGaku
				wJyoukenLog.HyokaGakuGokei = wJyoukenLog.KaiGakuGoukei + wJyoukenLog.Soneki
				wJyoukenLog.HoyuCnt = len(wHoyu)
				wJyoukenLogs = append(wJyoukenLogs, wJyoukenLog)
			}
			wHizuke = wKabuka.Hizuke
		}

		var wKaiTiming1 string
		wBuy = false
		if ((wKabuka.AveMEAN5Kairi.Float64 >= iJyouken.Kairi5.Float64 && wKabuka.AveMEAN5Kairi.Valid && iJyouken.Kairi5.Valid &&
			 wKabuka.AveMEAN25Kairi.Float64 <= iJyouken.Kairi25.Float64 && wKabuka.AveMEAN25Kairi.Valid &&iJyouken.Kairi25.Valid && iJyouken.KBN == "K") ||
			(wKabuka.AveMEAN5Kairi.Float64 <= iJyouken.Kairi5.Float64 && wKabuka.AveMEAN5Kairi.Valid && iJyouken.Kairi5.Valid &&
			 wKabuka.AveMEAN25Kairi.Float64 >= iJyouken.Kairi25.Float64 && wKabuka.AveMEAN25Kairi.Valid && iJyouken.Kairi25.Valid && iJyouken.KBN == "U")) &&
			wKabuka.OwariNe >= 50 {
			//買い方法は０成行のみ
			wUriNe = wKabuka.HazimariNeF1
			wKaiTiming1 = "N"
			wBuy = true
		}

		wBaibaiKekka, ok := wHoyu[wKabuka.MeigaraCd]
		if ok {
			//二日以上離れた場合、購入した情報はリセットする。
			if wBaibaiKekka.HizukeNum+2 < wKabuka.HizukeNum {
				wKoza += (wBaibaiKekka.KaiKabugaku + wBaibaiKekka.NaninKoza)
				delete(wHoyu, wBaibaiKekka.MeigaraCd)
				ok = false
			}
		}

		if ok {

			if wBuy {
				iJyouken.CntHoyutyu++
			}

			//終値で売り、ナンピン
			wKoza,wBaibaiKekka  =hanSel(wKoza,"O",iJyouken,wBaibaiKekka,wKabuka)
			wKoza,wBaibaiKekka = hanNanpin(wKoza,"O",iJyouken,wBaibaiKekka,wKabuka)
			//寄付きで売り、ナンピン
			wKoza,wBaibaiKekka  =hanSel(wKoza,"Y",iJyouken,wBaibaiKekka,wKabuka)
			wKoza,wBaibaiKekka = hanNanpin(wKoza,"Y",iJyouken,wBaibaiKekka,wKabuka)
			//指し値売り
			wKoza,wBaibaiKekka  =hanSel(wKoza,"S",iJyouken,wBaibaiKekka,wKabuka)

			if wBaibaiKekka.UriIdx > 0 {
				//売れた場合、
				wSimbaibaikekkaList = append(wSimbaibaikekkaList, wBaibaiKekka)
				delete(wHoyu, wBaibaiKekka.MeigaraCd)
			} else {
				//売れなかった場合
				wBaibaiKekka.UriKabuka = wKabuka.OwariNe
				wBaibaiKekka.UriKabugaku = wBaibaiKekka.UriKabuka  * float64(wBaibaiKekka.KaiKabusu)
				wBaibaiKekka.HizukeNum = wKabuka.HizukeNum
				if wBaibaiKekka.KaiIdxSl[3] > 0 {
					wBaibaiKekka.Day = wKabuka.HizukeNum - wBaibaiKekka.KaiIdxSl[3]
				} else if wBaibaiKekka.KaiIdxSl[2] > 0 {
					wBaibaiKekka.Day = wKabuka.HizukeNum - wBaibaiKekka.KaiIdxSl[2]
				} else if wBaibaiKekka.KaiIdxSl[0] > 0 {
					wBaibaiKekka.Day = wKabuka.HizukeNum - wBaibaiKekka.KaiIdxSl[1]
				} else {
					wBaibaiKekka.Day = wKabuka.HizukeNum - wBaibaiKekka.KaiIdxSl[0]
				}
				if wBaibaiKekka.KaiKabugaku > 0 {
					if iJyouken.KBN == "K" {
						wBaibaiKekka.Soneki = wBaibaiKekka.UriKabugaku - wBaibaiKekka.KaiKabugaku
					} else {
						wBaibaiKekka.Soneki = wBaibaiKekka.KaiKabugaku - wBaibaiKekka.UriKabugaku
					}
					wBaibaiKekka.SonekiRitu = (wBaibaiKekka.Soneki/wBaibaiKekka.KaiKabugaku ) * 100
				}
				wBaibaiKekka.DayAll = wKabuka.HizukeNum - wBaibaiKekka.KaiIdxSl[0]
				if wBaibaiKekka.MinKabuka > wKabuka.YasuNeF1 {
					wBaibaiKekka.MinKabuka = wKabuka.YasuNeF1
				}

				if wBaibaiKekka.MaxKabuka < wKabuka.TakaNeF1 {
					wBaibaiKekka.MaxKabuka = wKabuka.TakaNeF1
				}
				if iJyouken.KBN == "K" {
					wBaibaiKekka.MinKabukaRitu = ((wBaibaiKekka.MinKabuka - wBaibaiKekka.KaiKabuka)/wBaibaiKekka.KaiKabuka) * 100
					wBaibaiKekka.MaxKabukaRitu = ((wBaibaiKekka.MaxKabuka - wBaibaiKekka.KaiKabuka)/wBaibaiKekka.KaiKabuka) * 100
				} else {
					wBaibaiKekka.MinKabukaRitu = ((wBaibaiKekka.KaiKabuka - wBaibaiKekka.MinKabuka)/wBaibaiKekka.KaiKabuka) * 100
					wBaibaiKekka.MaxKabukaRitu = ((wBaibaiKekka.KaiKabuka - wBaibaiKekka.MaxKabuka)/wBaibaiKekka.KaiKabuka) * 100
				}
				wHoyu[wKabuka.MeigaraCd] = wBaibaiKekka				
			}
		} else {
			if wBuy {
				wSuu := com.GetKabuSuu(wKabuka.HazimariNeF1)
				wKonyu := wUriNe * float64(wSuu)
				wNanpinGaku := 0.0
				if wKoza >= wKonyu+wNanpinGaku || iJyouken.KozaMusi == 1 {
					var wSimbaibaikekka com.TsimbaibaikekkaGen
					wSimbaibaikekka.GenHizuke = iJyouken.GenHizuke
					wSimbaibaikekka.GenHizukeNum = iJyouken.GenHizukeNum
					wSimbaibaikekka.LogIdx = wCnt
					wSimbaibaikekka.MeigaraCd = wKabuka.MeigaraCd
					wSimbaibaikekka.KaiHizukeSl[0] = wKabuka.Hizuke
					wSimbaibaikekka.KaiIdxSl[0] = wKabuka.HizukeNum
					wSimbaibaikekka.KaiKabuka = wUriNe
					wSimbaibaikekka.KaiKabusu = wSuu
					wSimbaibaikekka.KaiKabugaku = wSimbaibaikekka.KaiKabuka * float64(wSimbaibaikekka.KaiKabusu)
					wSimbaibaikekka.UriKabugaku = wSimbaibaikekka.KaiKabugaku

					wSimbaibaikekka.KaiKabukaSl[0] = wSimbaibaikekka.KaiKabuka
					wSimbaibaikekka.KaiKabusuSl[0] = wSuu
					wSimbaibaikekka.KaiKabugakuSl[0] = wSimbaibaikekka.KaiKabukaSl[0] * float64(wSimbaibaikekka.KaiKabusuSl[0])
					wSimbaibaikekka.KaiTimingSl[0] = wKaiTiming1

					wSimbaibaikekka.NanpinYotei = wSimbaibaikekka.KaiKabuka * iJyouken.Nanpin
					wSimbaibaikekka.UriYotei = wSimbaibaikekka.KaiKabuka * iJyouken.UriRitu
					if iJyouken.UriRitu != iJyouken.UriSasiNeRitu {
						wSimbaibaikekka.UriSasiYotei = wSimbaibaikekka.KaiKabuka * iJyouken.UriSasiNeRitu
					}
					wSimbaibaikekka.MinKabuka = wKabuka.YasuNeF1
					wSimbaibaikekka.MaxKabuka = wKabuka.TakaNeF1

					if iJyouken.KBN == "K" {
						wSimbaibaikekka.MinKabukaRitu = ((wSimbaibaikekka.MinKabuka - wSimbaibaikekka.KaiKabuka)/wSimbaibaikekka.KaiKabuka) * 100
						wSimbaibaikekka.MaxKabukaRitu = ((wSimbaibaikekka.MaxKabuka - wSimbaibaikekka.KaiKabuka)/wSimbaibaikekka.KaiKabuka) * 100
					} else {
						wSimbaibaikekka.MinKabukaRitu = ((wSimbaibaikekka.KaiKabuka - wSimbaibaikekka.MinKabuka)/wSimbaibaikekka.KaiKabuka) * 100
						wSimbaibaikekka.MaxKabukaRitu = ((wSimbaibaikekka.KaiKabuka - wSimbaibaikekka.MaxKabuka)/wSimbaibaikekka.KaiKabuka) * 100
					}

					wSimbaibaikekka.NaninKoza = wNanpinGaku
					wSimbaibaikekka.HizukeNum = wKabuka.HizukeNum

					wSimbaibaikekka.TesuRyo += com.GetTesuRyo(iJyouken.KBN, wSimbaibaikekka.KaiKabugakuSl[0])

					wHoyu[wKabuka.MeigaraCd] = wSimbaibaikekka

					wKoza -= (wKonyu + wNanpinGaku)
					wCnt++
				} else {
					iJyouken.CntHusoku++
				}
			}
		}
	}

	if len(wHizuke) > 0 {
		var wJyoukenLog com.TJyokenLogGen
		wJyoukenLog.GenHizuke = iJyouken.GenHizuke
		wJyoukenLog.GenHizukeNum = iJyouken.GenHizukeNum
		wJyoukenLog.Koza = math.Round(wKoza)
		wJyoukenLog.Hizuke = wHizuke
		for _, wJyouken := range wHoyu {
			wJyoukenLog.KaiGaku += math.Round(wJyouken.KaiKabugaku)
			wJyoukenLog.HyokaGaku += math.Round(wJyouken.UriKabugaku)
		}
		if wJyoukenLog.KaiGaku > 0 {
			if iJyouken.KBN == "K" {
				wJyoukenLog.Soneki = wJyoukenLog.HyokaGaku - wJyoukenLog.KaiGaku
			} else {
				wJyoukenLog.Soneki = wJyoukenLog.KaiGaku - wJyoukenLog.HyokaGaku
			}
			wJyoukenLog.SonekiRitu = (wJyoukenLog.Soneki/wJyoukenLog.KaiGaku) * 100
			wJyoukenLog.KaiGakuGoukei = wJyoukenLog.Koza + wJyoukenLog.KaiGaku
			wJyoukenLog.HyokaGakuGokei = wJyoukenLog.KaiGakuGoukei + wJyoukenLog.Soneki
			wJyoukenLog.HoyuCnt = len(wHoyu)
			wJyoukenLogs = append(wJyoukenLogs, wJyoukenLog)
		}
	}

	for _, wSimbaibaikekka := range wSimbaibaikekkaList {
		if wSimbaibaikekka.KaiKabugaku > 0 {
			iJyouken.Cnt++
			iJyouken.HoyuDayAvg += float64(wSimbaibaikekka.Day)
			iJyouken.SonekiSum += wSimbaibaikekka.SonekiRitu * float64(wSimbaibaikekka.KaiKabusu / wSimbaibaikekka.KaiKabusuSl[0]) 
			if wSimbaibaikekka.SonekiRitu > 0 {
				iJyouken.SyoRitu += 100
			}
		}
	}

	if iJyouken.Cnt > 0 {
		iJyouken.HoyuDayAvg /= float64(iJyouken.Cnt)
		iJyouken.SonekiAve = iJyouken.SonekiSum / float64(iJyouken.Cnt)
		iJyouken.SyoRitu /= float64(iJyouken.Cnt)
	} else {
		iJyouken.HoyuDayAvg = 0.0
		iJyouken.SonekiAve = 0.0
		iJyouken.SyoRitu = 0.0
	}

	for _, wBaibaiKekka := range wHoyu {
		wKoza += math.Round(wBaibaiKekka.KaiKabugaku + wBaibaiKekka.Soneki)
		wSimbaibaikekkaList = append(wSimbaibaikekkaList, wBaibaiKekka)
	}

	iJyouken.Hyoka = wKoza / iJyouken.StaKoza * 100
	iJyouken.CntCancel = len(wHoyu)

	return iJyouken, wSimbaibaikekkaList, wJyoukenLogs
}



func setJyouken(iJyoukens []com.TJyoukenGen,
	iGenHizuke string,iGenHizukeNum int,
	iNoSta int,
	iNoEnd int,
	iNo int,
	iKeyNo int,
	iUriKai string,
	iK5Idx float64, iK25Idx float64,
	iWaitIdx int, iNanIdx float64, iUriIdx float64, iUriSasiNe float64, 
	iSKoza float64, iKozaOn int,
	iNanpinLvl int) (int, int, []com.TJyoukenGen) {


	//売り方法　0:終値で判定し、翌日寄り付きで売る。1：終値で売る、２：翌日寄り付きで売る、３：１と２

	//買い方法 0:成り行き
	kaiHohoIdx := 0
	//ナンピン方法も売り方法も３固定：寄付き、終値で売る
	nanHohoIdx := 3
	uriHohoIdx := 3

	for hoyuIdx := 10; hoyuIdx <= 80; hoyuIdx += 10 {
		if iNoSta <= iNo && iNo <= iNoEnd {
			var wJyouken com.TJyoukenGen
			wJyouken.GenHizuke = iGenHizuke
			wJyouken.GenHizukeNum = iGenHizukeNum
			wJyouken.ScenarioNo = iKeyNo
			wJyouken.KBN = iUriKai
			wJyouken.Kairi5.Float64 = iK5Idx
			wJyouken.Kairi5.Valid = true
			wJyouken.Kairi25.Float64 = iK25Idx
			wJyouken.Kairi25.Valid = true
			wJyouken.KaiHoho = kaiHohoIdx
			wJyouken.HoyuDay = float64(hoyuIdx)
			wJyouken.Nanpin = iNanIdx / 100
			wJyouken.NanpinKaiHoho = nanHohoIdx
			wJyouken.UriHoho = uriHohoIdx
			wJyouken.UriRitu = iUriIdx / 100
			wJyouken.UriSasiNeRitu = iUriSasiNe / 100
			wJyouken.SonekiSum = 0.0
			wJyouken.SonekiAve = 0.0
			wJyouken.StaKoza = iSKoza
			wJyouken.KozaMusi = iKozaOn
			wJyouken.NanpinLvl = iNanpinLvl
			iJyoukens = append(iJyoukens, wJyouken)
			iKeyNo++
			//fmt.Println(wJyouken)
		}
		iNo++
	}

	return iNo, iKeyNo, iJyoukens
}

func main() {

	var (
		aUriKai    = flag.String("UriKai", "K", "K:買い条件、U:売り条件")
		aDBStrring = flag.String("DBStrring", "kabu:kabukabu@tcp(localhost:3306)/kabu_test", "DB接続文字列")
		aDebug     = flag.Bool("Debug", false, "True:デバッグモード")

		aRunMode = flag.String("RunMode", "SIM", "SIM:シミュレーション,GEN:条件生成")

		aScenarioDB     = flag.String("ScenarioDB", "", "条件シナリオDBの後ろ")
		aScenarioDBInit = flag.Bool("ScenarioDBInit", false, "True:実行前にクリアする")

		aScenarioStaNo = flag.Int("ScenarioStaNo", -1, "-1:続きから生成,0以外:抽出開始NO")
		aScenarioAddNo = flag.Int("ScenarioAddNo", 1000000, "指定件数分生成する")

		aSimDB       = flag.String("SimDB", "Sim", "シミュレーション結果DBの後ろ")
		aSimDBInit   = flag.Bool("SimDBInit", false, "True:実行前にクリアする")
		aSimDBWhere  = flag.String("SimDBWhere", "LIMIT 100", "指定した抽出条件でシナリオDBから抽出しシミュレーションを実行")

		aKozaSta       = flag.Int("KozaSta", 1000000, "開始口座額")
		aNanpinLvl = flag.Int("NanpinLvl", 2, "ナンピンを行う段階")
		aKozaOn = flag.Int("KozaOn", 1, "口座を考慮するか")
	)
	flag.Parse()
	fmt.Printf("UriKai: %v\n", *aUriKai)
	fmt.Printf("DBStrring: %v\n", *aDBStrring)
	fmt.Printf("Debug: %v\n", *aDebug)
	fmt.Printf("RunMode: %v\n", *aRunMode)
	fmt.Printf("ScenarioDB: %v\n", *aScenarioDB)
	fmt.Printf("ScenarioDBInit: %v\n", *aScenarioDBInit)
	fmt.Printf("ScenarioStaNo: %v\n", *aScenarioStaNo)
	fmt.Printf("ScenarioAddNo: %v\n", *aScenarioAddNo)
	fmt.Printf("SimDB: %v\n", *aSimDB)
	fmt.Printf("SimDBInit: %v\n", *aSimDBInit)
	fmt.Printf("SimDBWhere: %v\n", *aSimDBWhere)
	fmt.Printf("KozaSta: %v\n", *aKozaSta)
	fmt.Printf("KozaOn: %v\n", *aKozaOn)
	fmt.Printf("NanpinLvl: %v\n", *aNanpinLvl)

	//:db、2:db2、3:db3
	//SW := ""
	//0:本番、1:テスト
	//wDEBUG := 0

	db, err := sql.Open("mysql", *aDBStrring)
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	var wSQL string
	var wKabukas VkabukaAvePool
	var ukabuka com.TkabukaAve
	var wTkabukaLast com.TkabukaLast

	wSQL = "SELECT MAX(Hizuke),MAX(HizukeNum) FROM tkabuka WHERE Hizuke is not null and HizukeNum is not null"
	rowLast, err := db.Query(wSQL)
	if err != nil {
		panic(err.Error())
	}
	defer rowLast.Close()

	rowLast.Next()
	errLast := rowLast.Scan(&wTkabukaLast.Hizuke, &wTkabukaLast.HizukeNum)
	if errLast != nil {
		panic(errLast.Error())
	}
	fmt.Println(wTkabukaLast)

	if *aDebug {
		wSQL = "SELECT MeigaraCd,Count(*) Cnt FROM tkabuka WHERE Hizuke >= DATE_ADD(CURDATE(),INTERVAL -14 MONTH) and MeigaraCd LIKE '9%' Group By MeigaraCd "
	} else {
		if *aUriKai == "K" {
			wSQL = "SELECT MeigaraCd,Count(*) Cnt FROM tkabuka WHERE Hizuke >= DATE_ADD(CURDATE(),INTERVAL -14 MONTH) and MeigaraCd REGEXP '^[0-9]+$' Group By MeigaraCd "
		} else {
			wSQL = "SELECT MeigaraCd,Count(*) Cnt FROM tkabuka WHERE Hizuke >= DATE_ADD(CURDATE(),INTERVAL -14 MONTH) and TaiSyaku = 1 and MeigaraCd REGEXP '^[0-9]+$' Group By MeigaraCd "
		}
	}
	rowSums, err := db.Query(wSQL)
	if err != nil {
		panic(err.Error())
	}
	defer rowSums.Close()


	var ukabukaGroup com.TkabukaGroup

	for rowSums.Next() {
		err := rowSums.Scan(&ukabukaGroup.MeigaraCd, &ukabukaGroup.Cnt)
		if err != nil {
			panic(err.Error())
		}
		fmt.Println(ukabukaGroup)

		wSQL = fmt.Sprintf("SELECT MeigaraCd,HizukeNum,Hizuke,HazimariNe,OwariNe,ZenOwariNe,HazimariNeF1,TakaNeF1,YasuNeF1,AveMEAN5Kairi,AveMEAN10Kairi,AveMEAN25Kairi,AveMEAN40Kairi,AveMEAN80Kairi,AveMEAN160Kairi  FROM tkabuka where MeigaraCd = '%s' ORDER BY Hizuke DESC", ukabukaGroup.MeigaraCd)

		rows, err := db.Query(wSQL)
		if err != nil {
			panic(err.Error())
		}

		var wKabukasTmp VkabukaAvePool

		//一件目
		for rows.Next() {

			err := rows.Scan(&ukabuka.MeigaraCd, &ukabuka.HizukeNum, &ukabuka.Hizuke, &ukabuka.HazimariNe, &ukabuka.OwariNe, &ukabuka.ZenOwariNe, &ukabuka.HazimariNeF1, &ukabuka.TakaNeF1, &ukabuka.YasuNeF1,
				 &ukabuka.AveMEAN5Kairi, &ukabuka.AveMEAN10Kairi,&ukabuka.AveMEAN25Kairi,&ukabuka.AveMEAN40Kairi,&ukabuka.AveMEAN80Kairi,&ukabuka.AveMEAN160Kairi)
			if err != nil {
				panic(err.Error())
			}
			defer rows.Close()
			wAns, wKabuka := com.ChkKabukaAve(ukabuka)

			if wAns {
				wIdx := len(wKabukasTmp) - 1
				if wIdx >= 0 {
					wHenka := (wKabuka.OwariNe/wKabukasTmp[wIdx].OwariNe - 1) * 100
					if wHenka >= 50 || wHenka <= -50 {
						fmt.Println("Err_OwarineStep:", wKabuka.OwariNe, "=>", wKabukasTmp[wIdx].OwariNe)
						break
					}
				}
				wKabukasTmp = append(wKabukasTmp, wKabuka)
			}
		}
		rows.Close()

		for _, wKabuka := range wKabukasTmp {
			wKabukas = append(wKabukas, wKabuka)
		}
		fmt.Println(ukabukaGroup.MeigaraCd, ":Cnt=", len(wKabukasTmp))
	}

	fmt.Println("TotalCnt=", len(wKabukas))
	sort.Sort(VkabukaAvePool(wKabukas))

	var wJyoukens []com.TJyoukenGen
	var wHyokaMax float64
	var wKeyNo int

	if *aRunMode == "GEN" {
		wNo := 0
		if *aScenarioDBInit == false {
			wKeyNo, wHyokaMax = com.ScenarioDBNoGen(db, *aScenarioDB,wTkabukaLast.Hizuke.String,wTkabukaLast.HizukeNum.Int64)
		}
		SKoza := float64(*aKozaSta)
		waitIdx := 0

		wNoSta := 0
		if *aScenarioStaNo < 0 {
			wNoSta = wKeyNo
		} else {
			wNoSta = *aScenarioStaNo
		}

		wNoEnd := wNoSta + *aScenarioAddNo

		if *aUriKai == "K" {
			for k25Idx := -40.0; k25Idx <= float64(-5); k25Idx += 2.5 {
				k5End := 5.0
				k5Step := math.Ceil((k5End - k25Idx) / 10)  
				for k5Idx := k25Idx; k5Idx <= k5End; k5Idx += k5Step {
					for nanIdx := 80.0; nanIdx <= 100.0; nanIdx += 2.5 {
						for uriIdx := 105.0; uriIdx <= 125.0; uriIdx += 2.5 {
							for uriSasiIdx := uriIdx; uriSasiIdx <= uriIdx + 20; uriSasiIdx += 2.5 {
								wNo, wKeyNo, wJyoukens = setJyouken(wJyoukens,
								wTkabukaLast.Hizuke.String,int(wTkabukaLast.HizukeNum.Int64),
								 wNoSta, wNoEnd, wNo, wKeyNo, *aUriKai, k5Idx, k25Idx, waitIdx, nanIdx, uriIdx,uriSasiIdx, SKoza,*aKozaOn,
								*aNanpinLvl)
							}
						}
					}
				}
			}
		} else {
			for k25Idx := 40.0; k25Idx >= float64(5); k25Idx -= 2.5 {
				k5End := -5.0
				k5Step := math.Ceil((k25Idx - k5End) / 10)  
				for k5Idx := k25Idx; k5Idx >= k5End; k5Idx -= k5Step {
					for nanIdx := 120.0; nanIdx >= 100.0; nanIdx -= 2.5 {
						for uriIdx := 95.0; uriIdx >= 75.0; uriIdx -= 2.5 {
							for uriSasiIdx := uriIdx; uriSasiIdx >= uriIdx - 20; uriSasiIdx -= 2.5 {
								wNo, wKeyNo, wJyoukens = setJyouken(wJyoukens,
								wTkabukaLast.Hizuke.String,int(wTkabukaLast.HizukeNum.Int64),
								 wNoSta, wNoEnd, wNo, wKeyNo, *aUriKai, k5Idx, k25Idx, waitIdx, nanIdx, uriIdx,uriSasiIdx, SKoza,*aKozaOn,
								*aNanpinLvl)
							}
						}
					}
				}
			}
		}
	} else if *aRunMode == "SIM" {
		if *aSimDBInit {
			wSQL = fmt.Sprintf("SELECT * FROM tjyouken%s ORDER BY KAIRI5,KAIRI25,Hyoka desc", *aScenarioDB)
		} else {
			wSQL = fmt.Sprintf("SELECT * FROM tjyouken%s ORDER BY KAIRI5,KAIRI25,Hyoka desc", *aSimDB)
		}
		rows, err := db.Query(wSQL)
		if err != nil {
			panic(err.Error())
		}
		var wKairi5 float64
		var wKairi25 float64
		for rows.Next() {
			wVJyouken := com.GetVJyokenGenScan(rows)
			wAns, wJyouken := com.ChkJyoukenGen(wVJyouken)
			if wAns {
				if wKairi5 != wVJyouken.Kairi5.Float64 || wKairi25 != wVJyouken.Kairi25.Float64 {
					wKairi5 = wVJyouken.Kairi5.Float64
					wKairi25 = wVJyouken.Kairi25.Float64
					fmt.Println(wJyouken)

					wJyouken.Cnt = 0
					wJyouken.SonekiSum = 0
					wJyouken.SonekiAve = 0
					wJyouken.HoyuDayAvg = 0
					wJyouken.Hyoka = 0
					wJyouken.CntHoyutyu = 0
					wJyouken.CntHusoku = 0
					wJyouken.CntCancel = 0
					wJyouken.SyoRitu = 0

					wJyoukens = append(wJyoukens, wJyouken)
				}
			}
		}
		rows.Close()
	}

	var delTable []string

	if *aRunMode == "GEN" {
		if *aScenarioDBInit {
			delTable = append(delTable, "tjyouken")
		}
	} else if *aRunMode == "SIM" {
		//if *aSimDBInit {
		delTable = append(delTable, "tjyouken")
		//copyJyoken(db, *aScenarioDB, *aSimDB)
		//}
	}
	delTable = append(delTable, "tsimbaibaikekka")
	delTable = append(delTable, "tjyoukenlog")
	if *aRunMode == "GEN" {
		com.InitDbGen(db, delTable, *aScenarioDB,wTkabukaLast.Hizuke.String,wTkabukaLast.HizukeNum.Int64)
	} else if *aRunMode == "SIM" {
		com.InitDbGen(db, delTable, *aSimDB,wTkabukaLast.Hizuke.String,wTkabukaLast.HizukeNum.Int64)
	}

	var wg sync.WaitGroup
	var cpus int
	var OutDb string

	if *aRunMode == "GEN" {
		OutDb = *aScenarioDB
	} else if *aRunMode == "SIM" {
		OutDb = *aSimDB
	}

	if *aDebug {
		cpus = 1
	} else {
		cpus = int(float64(runtime.NumCPU()) * 0.75)
	}
	semaphore := make(chan int, cpus)

	for i, wJyouken := range wJyoukens {
		wg.Add(1)
		go func(wJyouken com.TJyoukenGen) {
			defer wg.Done()
			semaphore <- i
			wJyoukenKekka, wSimbaibaikekkaList, wJyoukenLogs := runBuySim(*aScenarioDB, wJyouken, wKabukas)
			fmt.Println(wJyoukenKekka)
			if wJyoukenKekka.Hyoka > 100 || *aRunMode == "SIM" {
				com.InsertJyoukenGen(db,OutDb,wJyoukenKekka)
				if wHyokaMax < wJyoukenKekka.Hyoka || *aRunMode == "SIM" {
					wHyokaMax = wJyoukenKekka.Hyoka
					for j, wSimbaibaikekka := range wSimbaibaikekkaList {
						//if (wSimbaibaikekkaList[i].UriIdx >0){
						wSimbaibaikekka.LogIdx = j
						wSimbaibaikekka.ScenarioNo = wJyoukenKekka.ScenarioNo
						wSimbaibaikekka.Kbn = "00"
						com.InsertSimbaibaikekkaGen(db,OutDb,wSimbaibaikekka)
					}

					for _, wJyoukenLog := range wJyoukenLogs {
						com.InsertJyokenLogGen(db,OutDb,wJyoukenLog)
					}
				}
			}
			<-semaphore
		}(wJyouken)

	}
	wg.Wait()
}
