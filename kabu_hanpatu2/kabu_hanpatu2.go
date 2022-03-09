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
type VkabukaPool []com.Vkabuka
// 以下インタフェースを満たす

func (p VkabukaPool) Len() int {
	return len(p)
}

func (p VkabukaPool) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p VkabukaPool) Less(i, j int) bool {
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


func runBuySim(iSW string, iJyouken com.TJyouken, iKabukas []com.Vkabuka) (com.TJyouken, []com.Tsimbaibaikekka, []com.TJyokenLog) {

	var wSimbaibaikekkaList []com.Tsimbaibaikekka
	var wJyoukenLogs []com.TJyokenLog

	wHizuke := ""
	wKoza := iJyouken.StaKoza

	//wKozaOff := false        //true:口座無視、false：口座残高確認
	//wNanpinKozaOff := false  //true:ナンピンを信用で買う false:現金
	//wNanpinKozaZanOn := true //True:現金でナンピン用のお金を残す

	wHoyu := make(map[string]com.Tsimbaibaikekka)
	wCnt := 0
	var wBuy bool
	wUriNe := 0.0
	for _, wKabuka := range iKabukas {

		if wHizuke != wKabuka.Hizuke {
			if len(wHizuke) > 0 {
				var wJyoukenLog com.TJyokenLog
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
						wJyoukenLog.SonekiRitu = (wJyoukenLog.HyokaGaku/wJyoukenLog.KaiGaku - 1) * 100
					} else {
						wJyoukenLog.Soneki = wJyoukenLog.KaiGaku - wJyoukenLog.HyokaGaku
						wJyoukenLog.SonekiRitu = (wJyoukenLog.KaiGaku/wJyoukenLog.HyokaGaku - 1) * 100
					}
				}
				wJyoukenLog.KaiGakuGoukei = wJyoukenLog.Koza + wJyoukenLog.KaiGaku
				wJyoukenLog.HyokaGakuGokei = wJyoukenLog.KaiGakuGoukei + wJyoukenLog.Soneki
				wJyoukenLog.HoyuCnt = len(wHoyu)
				wJyoukenLogs = append(wJyoukenLogs, wJyoukenLog)
			}
			wHizuke = wKabuka.Hizuke
		}

		//	if wOwariNe {
		//		wUriNe = wKabuka.OwariNe
		//	} else {
		//			wUriNe = wKabuka.HazimariNeF1
		//		}

		var wKaiTiming1 string
		wBuy = false
		if ((wKabuka.MEAN5Kairi >= iJyouken.Kairi5 && wKabuka.MEAN25Kairi <= iJyouken.Kairi25 && iJyouken.KBN == "K") ||
			(wKabuka.MEAN5Kairi <= iJyouken.Kairi5 && wKabuka.MEAN25Kairi >= iJyouken.Kairi25 && iJyouken.KBN == "U")) &&
			wKabuka.OwariNe >= 50 {
			wUriNe = wKabuka.HazimariNeF1
			if iJyouken.KaiHoho == 0 {
				wKaiTiming1 = "N"
				wBuy = true
			} else {
				if iJyouken.KBN == "K" {
					if wKabuka.OwariNe > wKabuka.HazimariNeF1 {
						wKaiTiming1 = "N"
						wBuy = true
					} else if wKabuka.OwariNe <= wKabuka.TakaNeF1 && wKabuka.OwariNe >= wKabuka.YasuNeF1 {
						wKaiTiming1 = "S"
						wBuy = true
						wUriNe = wKabuka.OwariNe
					}
				} else {
					if wKabuka.OwariNe < wKabuka.HazimariNeF1 {
						wKaiTiming1 = "N"
						wBuy = true
					} else if wKabuka.OwariNe <= wKabuka.TakaNeF1 && wKabuka.OwariNe >= wKabuka.YasuNeF1 {
						wKaiTiming1 = "S"
						wBuy = true
						wUriNe = wKabuka.OwariNe
					}
				}
			}
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

			wBaibaiKekka.UriKabuka = wKabuka.OwariNe
			wBaibaiKekka.UriKabugaku = wKabuka.OwariNe * float64(wBaibaiKekka.KaiKabusu)
			wBaibaiKekka.HizukeNum = wKabuka.HizukeNum
			if wBaibaiKekka.KaiIdx3 > 0 {
				wBaibaiKekka.Day = wKabuka.HizukeNum - wBaibaiKekka.KaiIdx3
			} else if wBaibaiKekka.KaiIdx2 > 0 {
				wBaibaiKekka.Day = wKabuka.HizukeNum - wBaibaiKekka.KaiIdx2
			} else {
				wBaibaiKekka.Day = wKabuka.HizukeNum - wBaibaiKekka.KaiIdx
			}
			if wBaibaiKekka.KaiKabugaku > 0 {
				if iJyouken.KBN == "K" {
					wBaibaiKekka.Soneki = wBaibaiKekka.UriKabugaku - wBaibaiKekka.KaiKabugaku
					wBaibaiKekka.SonekiRitu = (wBaibaiKekka.UriKabugaku/wBaibaiKekka.KaiKabugaku - 1) * 100
				} else {
					wBaibaiKekka.Soneki = wBaibaiKekka.KaiKabugaku - wBaibaiKekka.UriKabugaku
					wBaibaiKekka.SonekiRitu = (wBaibaiKekka.KaiKabugaku/wBaibaiKekka.UriKabugaku - 1) * 100
				}
			}

			wSelOk := true
			//指し値
			wNanpinBuy := false
			//ナンピン条件
			wKaiKingaku := wBaibaiKekka.KaiKabuka * iJyouken.Nanpin
			wTiming := ""
			if iJyouken.Nanpin != 1 {
				if iJyouken.NanpinKaiHoho == 0 {
					if (wKaiKingaku >= wKabuka.OwariNe && iJyouken.KBN == "K") ||
						(wKaiKingaku <= wKabuka.OwariNe && iJyouken.KBN == "U") {
						wNanpinBuy = true
						wUriNe = wKabuka.HazimariNeF1
						wTiming = "Y"
					}
				}

				if wNanpinBuy == false && (iJyouken.NanpinKaiHoho == 1 || iJyouken.NanpinKaiHoho == 3) {
					if (wKaiKingaku >= wKabuka.OwariNe && iJyouken.KBN == "K") ||
						(wKaiKingaku <= wKabuka.OwariNe && iJyouken.KBN == "U") {
						wNanpinBuy = true
						wUriNe = wKabuka.OwariNe
						wTiming = "O"
					}
				}
				if wNanpinBuy == false && (iJyouken.NanpinKaiHoho == 2 || iJyouken.NanpinKaiHoho == 3) {
					if (wKaiKingaku >= wKabuka.HazimariNeF1 && iJyouken.KBN == "K") ||
						(wKaiKingaku <= wKabuka.HazimariNeF1 && iJyouken.KBN == "U") {
						wNanpinBuy = true
						wUriNe = wKabuka.HazimariNeF1
						wTiming = "Y"
					}
				}
			}

			if wNanpinBuy {
				wKonyu := wUriNe * float64(wBaibaiKekka.KaiKabusu)
				if iJyouken.KozaMusi == 0 {
					if wKoza < wKonyu {
						if wBaibaiKekka.KaiIdx2 == 0 && 1 <= iJyouken.NanpinLvl {
							wBaibaiKekka.KaiHusoku2++
						} else {
							if wBaibaiKekka.KaiIdx3 == 0 && 2 <= iJyouken.NanpinLvl {
								wBaibaiKekka.KaiHusoku3++
							}
						}
					}
				}
				if wKoza >= wKonyu || iJyouken.KozaMusi == 1  {
					if wBaibaiKekka.KaiIdx2 == 0 && 1 <= iJyouken.NanpinLvl {
						//買い増し
						wBaibaiKekka.KaiIdx2 = wKabuka.HizukeNum
						wBaibaiKekka.KaiHizuke2 = wKabuka.Hizuke
						wBaibaiKekka.KaiKabuka2 = wUriNe
						wBaibaiKekka.KaiKabusu2 = wBaibaiKekka.KaiKabusu
						wBaibaiKekka.KaiKabuka = (wBaibaiKekka.KaiKabuka1 + wBaibaiKekka.KaiKabuka2) / 2
						wBaibaiKekka.KaiKabugaku2 = wBaibaiKekka.KaiKabuka2 * float64(wBaibaiKekka.KaiKabusu2)
						wBaibaiKekka.KaiTiming2 = wTiming
						wBaibaiKekka.KaiKabugaku += wBaibaiKekka.KaiKabugaku2
						wBaibaiKekka.KaiKabusu += wBaibaiKekka.KaiKabusu2
						wBaibaiKekka.NanpinYotei = wBaibaiKekka.KaiKabuka * iJyouken.Nanpin
						wBaibaiKekka.UriYotei = wBaibaiKekka.KaiKabuka * iJyouken.UriRitu

						wBaibaiKekka.TesuRyo += com.GetTesuRyo(iJyouken.KBN, wBaibaiKekka.KaiKabugaku2)

						wKoza -= wKonyu

						wSelOk = false
					} else if wBaibaiKekka.KaiIdx3 == 0 && 2 <= iJyouken.NanpinLvl {
						//買い増し
						wBaibaiKekka.KaiIdx3 = wKabuka.HizukeNum
						wBaibaiKekka.KaiHizuke3 = wKabuka.Hizuke
						wBaibaiKekka.KaiKabuka3 = wUriNe
						wBaibaiKekka.KaiKabusu3 = wBaibaiKekka.KaiKabusu
						wBaibaiKekka.KaiKabugaku3 = wBaibaiKekka.KaiKabuka3 * float64(wBaibaiKekka.KaiKabusu3)
						wBaibaiKekka.KaiTiming3 = wTiming

						wBaibaiKekka.KaiKabuka = (wBaibaiKekka.KaiKabuka + wBaibaiKekka.KaiKabuka3) / 2
						wBaibaiKekka.KaiKabugaku += wBaibaiKekka.KaiKabugaku3
						wBaibaiKekka.KaiKabusu += wBaibaiKekka.KaiKabusu3
						wBaibaiKekka.NanpinYotei = 0
						wBaibaiKekka.UriYotei = wBaibaiKekka.KaiKabuka * iJyouken.UriRitu

						wBaibaiKekka.TesuRyo += com.GetTesuRyo(iJyouken.KBN, wBaibaiKekka.KaiKabugaku3)

						wKoza -= wKonyu

						wSelOk = false
					}
				}
				if wSelOk == false {
					//売り判定してはいけないということは、買い増しがあった。最大と最小をリセットする。
					wBaibaiKekka.MinKabuka = wKabuka.YasuNeF1
					wBaibaiKekka.MaxKabuka = wKabuka.TakaNeF1
				}
			}

			if wSelOk {
				//売り判定
				wSelOk = false
				wTiming := ""

				if iJyouken.UriHoho == 0 {
					if (wBaibaiKekka.UriYotei <= wKabuka.OwariNe && iJyouken.KBN == "K") ||
						(wBaibaiKekka.UriYotei >= wKabuka.OwariNe && iJyouken.KBN == "U") {
						wSelOk = true
						wUriNe = wKabuka.HazimariNeF1
						wTiming = "Y"
					}
				}

				if wSelOk == false && (iJyouken.UriHoho == 1 || iJyouken.UriHoho == 3) {
					if (wBaibaiKekka.UriYotei <= wKabuka.OwariNe && iJyouken.KBN == "K") ||
						(wBaibaiKekka.UriYotei >= wKabuka.OwariNe && iJyouken.KBN == "U") {
						wSelOk = true
						wUriNe = wKabuka.OwariNe
						wTiming = "O"
					}
				}

				if wSelOk == false && (iJyouken.UriHoho == 2 || iJyouken.UriHoho == 3) {
					if (wBaibaiKekka.UriYotei <= wKabuka.HazimariNeF1 && iJyouken.KBN == "K") ||
						(wBaibaiKekka.UriYotei >= wKabuka.HazimariNeF1 && iJyouken.KBN == "U") {
						wSelOk = true
						wUriNe = wKabuka.HazimariNeF1
						wTiming = "Y"
					}
				}

				if wSelOk == false && wKabuka.HizukeNum-wBaibaiKekka.KaiIdx >= int(iJyouken.HoyuDay) {
					wSelOk = true
					wUriNe = wKabuka.HazimariNeF1
					wTiming = "Y"
				}

				if wSelOk {
					wBaibaiKekka.UriIdx = wKabuka.HizukeNum
					wBaibaiKekka.UriHizuke = wKabuka.Hizuke
					wBaibaiKekka.DayAll = wBaibaiKekka.UriIdx - wBaibaiKekka.KaiIdx
					wBaibaiKekka.UriKabuka = wUriNe
					wBaibaiKekka.UriKabugaku = wBaibaiKekka.UriKabuka * float64(wBaibaiKekka.KaiKabusu)
					wBaibaiKekka.UriTiming = wTiming

					if iJyouken.KBN == "U" {
						var wDay int
						wDay = wBaibaiKekka.UriIdx - wBaibaiKekka.KaiIdx
						wBaibaiKekka.Kinri = com.GetKinri(iJyouken.KBN, wBaibaiKekka.KaiKabugaku1, wDay)
						if wBaibaiKekka.KaiIdx2 > 0 {
							wDay = wBaibaiKekka.UriIdx - wBaibaiKekka.KaiIdx2
							wBaibaiKekka.Kinri += com.GetKinri(iJyouken.KBN, wBaibaiKekka.KaiKabugaku2, wDay)
						}
						if wBaibaiKekka.KaiIdx3 > 0 {
							wDay = wBaibaiKekka.UriIdx - wBaibaiKekka.KaiIdx3
							wBaibaiKekka.Kinri += com.GetKinri(iJyouken.KBN, wBaibaiKekka.KaiKabugaku3, wDay)
						}
					}
					wBaibaiKekka.TesuRyo += com.GetTesuRyo(iJyouken.KBN, wBaibaiKekka.UriKabugaku)

					if wBaibaiKekka.KaiKabugaku > 0 {
						if iJyouken.KBN == "K" {
							wBaibaiKekka.Soneki = wBaibaiKekka.UriKabugaku - wBaibaiKekka.KaiKabugaku - wBaibaiKekka.TesuRyo - wBaibaiKekka.Kinri
							wBaibaiKekka.SonekiRitu = (wBaibaiKekka.UriKabugaku/wBaibaiKekka.KaiKabugaku - 1) * 100
						} else {
							wBaibaiKekka.Soneki = wBaibaiKekka.KaiKabugaku - wBaibaiKekka.UriKabugaku - wBaibaiKekka.TesuRyo - wBaibaiKekka.Kinri
							wBaibaiKekka.SonekiRitu = (wBaibaiKekka.KaiKabugaku/wBaibaiKekka.UriKabugaku - 1) * 100
						}
					}

					wSimbaibaikekkaList = append(wSimbaibaikekkaList, wBaibaiKekka)
					delete(wHoyu, wBaibaiKekka.MeigaraCd)

					if iJyouken.KBN == "K" {
						wKoza += (wBaibaiKekka.UriKabugaku + wBaibaiKekka.NaninKoza)
					} else {
						wKoza += (wBaibaiKekka.KaiKabugaku + wBaibaiKekka.KaiKabugaku - wBaibaiKekka.UriKabugaku + wBaibaiKekka.NaninKoza)
					}
				}
			}
			if wSelOk == false {
				wBaibaiKekka.DayAll = wKabuka.HizukeNum - wBaibaiKekka.KaiIdx
				if wBaibaiKekka.MinKabuka > wKabuka.YasuNeF1 {
					wBaibaiKekka.MinKabuka = wKabuka.YasuNeF1
				}

				if wBaibaiKekka.MaxKabuka < wKabuka.TakaNeF1 {
					wBaibaiKekka.MaxKabuka = wKabuka.TakaNeF1
				}
				if iJyouken.KBN == "K" {
					wBaibaiKekka.MinKabukaRitu = (wBaibaiKekka.MinKabuka/wBaibaiKekka.KaiKabuka - 1) * 100
					wBaibaiKekka.MaxKabukaRitu = (wBaibaiKekka.MaxKabuka/wBaibaiKekka.KaiKabuka - 1) * 100
				} else {
					wBaibaiKekka.MinKabukaRitu = (wBaibaiKekka.KaiKabuka/wBaibaiKekka.MinKabuka - 1) * 100
					wBaibaiKekka.MaxKabukaRitu = (wBaibaiKekka.KaiKabuka/wBaibaiKekka.MaxKabuka - 1) * 100
				}
				wHoyu[wKabuka.MeigaraCd] = wBaibaiKekka
			}
		} else {
			if wBuy {
				wSuu := com.GetKabuSuu(wKabuka.HazimariNeF1)
				wKonyu := wUriNe * float64(wSuu)
				wNanpinGaku := 0.0
				if wKoza >= wKonyu+wNanpinGaku || iJyouken.KozaMusi == 1 {
					var wSimbaibaikekka com.Tsimbaibaikekka
					wSimbaibaikekka.LogIdx = wCnt
					wSimbaibaikekka.MeigaraCd = wKabuka.MeigaraCd
					wSimbaibaikekka.Hizuke = wKabuka.Hizuke
					wSimbaibaikekka.KaiIdx = wKabuka.HizukeNum
					wSimbaibaikekka.KaiKabuka = wUriNe
					wSimbaibaikekka.KaiKabusu = wSuu
					wSimbaibaikekka.KaiKabugaku = wSimbaibaikekka.KaiKabuka * float64(wSimbaibaikekka.KaiKabusu)
					wSimbaibaikekka.UriKabugaku = wSimbaibaikekka.KaiKabugaku

					wSimbaibaikekka.KaiKabuka1 = wSimbaibaikekka.KaiKabuka
					wSimbaibaikekka.KaiKabusu1 = wSuu
					wSimbaibaikekka.KaiKabugaku1 = wSimbaibaikekka.KaiKabuka1 * float64(wSimbaibaikekka.KaiKabusu1)
					wSimbaibaikekka.KaiTiming1 = wKaiTiming1

					wSimbaibaikekka.NanpinYotei = wSimbaibaikekka.KaiKabuka * iJyouken.Nanpin
					wSimbaibaikekka.UriYotei = wSimbaibaikekka.KaiKabuka * iJyouken.UriRitu
					wSimbaibaikekka.MinKabuka = wKabuka.YasuNeF1
					wSimbaibaikekka.MaxKabuka = wKabuka.TakaNeF1

					if iJyouken.KBN == "K" {
						wSimbaibaikekka.MinKabukaRitu = (wSimbaibaikekka.MinKabuka/wSimbaibaikekka.KaiKabuka - 1) * 100
						wSimbaibaikekka.MaxKabukaRitu = (wSimbaibaikekka.MaxKabuka/wSimbaibaikekka.KaiKabuka - 1) * 100
					} else {
						wSimbaibaikekka.MinKabukaRitu = (wSimbaibaikekka.KaiKabuka/wSimbaibaikekka.MinKabuka - 1) * 100
						wSimbaibaikekka.MaxKabukaRitu = (wSimbaibaikekka.KaiKabuka/wSimbaibaikekka.MaxKabuka - 1) * 100
					}

					wSimbaibaikekka.NaninKoza = wNanpinGaku
					wSimbaibaikekka.HizukeNum = wKabuka.HizukeNum

					wSimbaibaikekka.TesuRyo += com.GetTesuRyo(iJyouken.KBN, wSimbaibaikekka.KaiKabugaku1)

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
		var wJyoukenLog com.TJyokenLog
		wJyoukenLog.Koza = math.Round(wKoza)
		wJyoukenLog.Hizuke = wHizuke
		for _, wJyouken := range wHoyu {
			wJyoukenLog.KaiGaku += math.Round(wJyouken.KaiKabugaku)
			wJyoukenLog.HyokaGaku += math.Round(wJyouken.UriKabugaku)
		}
		if wJyoukenLog.KaiGaku > 0 {
			if iJyouken.KBN == "K" {
				wJyoukenLog.Soneki = wJyoukenLog.HyokaGaku - wJyoukenLog.KaiGaku
				wJyoukenLog.SonekiRitu = (wJyoukenLog.HyokaGaku/wJyoukenLog.KaiGaku - 1) * 100
			} else {
				wJyoukenLog.Soneki = wJyoukenLog.KaiGaku - wJyoukenLog.HyokaGaku
				wJyoukenLog.SonekiRitu = (wJyoukenLog.KaiGaku/wJyoukenLog.HyokaGaku - 1) * 100
			}
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
			iJyouken.SonekiSum += wSimbaibaikekka.SonekiRitu
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



func setJyouken(iJyoukens []com.TJyouken,
	iNoSta int,
	iNoEnd int,
	iNo int,
	iKeyNo int,
	iUriKai string,
	iK5Idx float64, iK25Idx float64,
	iWaitIdx int, iNanIdx float64, iUriIdx float64, 
	iSKoza float64, iKozaOn int,
	iNanpinLvl int,
	iKaiSasiNe bool,
	iNanpinOwariNe bool, iNanpinYoriNe bool, iUriOwariNe bool, iUriYoriNe bool) (int, int, []com.TJyouken) {

	var kaiHohoIdx int
	var nanHohoIdx int
	var uriHohoIdx int

	if iKaiSasiNe {
		kaiHohoIdx = 1
	} else {
		kaiHohoIdx = 0
	}

	if iNanpinOwariNe {
		if iNanpinYoriNe {
			nanHohoIdx = 3
		} else {
			nanHohoIdx = 1
		}
	} else {
		if iNanpinYoriNe {
			nanHohoIdx = 2
		} else {
			nanHohoIdx = 0
		}
	}

	if iUriOwariNe {
		if iUriYoriNe {
			uriHohoIdx = 3
		} else {
			uriHohoIdx = 1
		}
	} else {
		if iUriYoriNe {
			uriHohoIdx = 2
		} else {
			uriHohoIdx = 0
		}
	}

	for hoyuIdx := 10; hoyuIdx <= 80; hoyuIdx += 10 {
		//for KozaMusiIdx := 0; KozaMusiIdx <= 1; KozaMusiIdx += 1 {
		//if KozaMusiIdx == 0 {
		//	wNanpinMax = 1
		//wNanpinMax = 1
		//	wNanpinKozaMusiIdxMax = 0
		//wNanpinKozaMusiIdxMax = 1
		//}
		if iNoSta <= iNo && iNo <= iNoEnd {
			var wJyouken com.TJyouken
			wJyouken.ScenarioNo = iKeyNo
			wJyouken.KBN = iUriKai
			wJyouken.Kairi5 = iK5Idx
			wJyouken.Kairi25 = iK25Idx
			wJyouken.KaiHoho = kaiHohoIdx
			wJyouken.HoyuDay = float64(hoyuIdx)
			wJyouken.Nanpin = iNanIdx / 100
			wJyouken.NanpinKaiHoho = nanHohoIdx
			wJyouken.UriHoho = uriHohoIdx
			wJyouken.UriRitu = iUriIdx / 100
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
		aAveNeOn    = flag.Bool("AveNeOn", false, "true:平均値で乖離率判定")
		aDBStrring = flag.String("DBStrring", "kabu:kabukabu@tcp(abee-t20:3306)/kabu_test", "DB接続文字列")
		aDebug     = flag.Bool("Debug", false, "True:デバッグモード")

		aRunMode = flag.String("RunMode", "SIM", "SIM:シミュレーション,GEN:条件生成")

		aScenarioDB     = flag.String("ScenarioDB", "", "条件シナリオDBの後ろ")
		aScenarioDBInit = flag.Bool("ScenarioDBInit", false, "True:実行前にクリアする")

		aScenarioStaNo = flag.Int("ScenarioStaNo", -1, "-1:続きから生成,0以外:抽出開始NO")
		aScenarioAddNo = flag.Int("ScenarioAddNo", 100000, "指定件数分生成する")

		aSimDB       = flag.String("SimDB", "Sim", "シミュレーション結果DBの後ろ")
		aSimDBInit   = flag.Bool("SimDBInit", false, "True:実行前にクリアする")
		aSimDBWhere  = flag.String("SimDBWhere", "LIMIT 100", "指定した抽出条件でシナリオDBから抽出しシミュレーションを実行")

		aKozaSta       = flag.Int("KozaSta", 1000000, "開始口座額")
		aNanpinLvl = flag.Int("NanpinLvl", 2, "ナンピンを行う段階")
		aKozaOn = flag.Int("KozaOn", 1, "口座を考慮するか")
		aKaiSasiNe     = flag.Bool("KaiSasiNe", false, "True:買いを指し値で行う")
		aNanpinOwariNe = flag.Bool("NanpinOwariNe", false, "True:ナンピンを終値で指し値買いする")
		aNanpinYoriNe  = flag.Bool("NanpinYoriNe", false, "True:ナンピンを寄付きで指し値買いする")
		aUriOwariNe    = flag.Bool("UriOwariNe", false, "True:売りを終値で指し値買いする")
		aUriYoriNe     = flag.Bool("UriYoriNe", false, "True:売りを寄付きで指し値買いする")
	)
	flag.Parse()
	fmt.Printf("UriKai: %v\n", *aUriKai)
	fmt.Printf("AveNeOn: %v\n", *aAveNeOn)
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
	fmt.Printf("KaiSasiNe: %v\n", *aKaiSasiNe)
	fmt.Printf("NanpinOwariNe: %v\n", *aNanpinOwariNe)
	fmt.Printf("NanpinYoriNe: %v\n", *aNanpinYoriNe)
	fmt.Printf("UriOwariNe: %v\n", *aUriOwariNe)
	fmt.Printf("UriYoriNe: %v\n", *aUriYoriNe)

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
	var wKabukas VkabukaPool
	var ukabuka com.Tkabuka

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

		if (*aAveNeOn){
			wSQL = fmt.Sprintf("SELECT MeigaraCd,HizukeNum,Hizuke,HazimariNe,OwariNe,ZenOwariNe,HazimariNeF1,TakaNeF1,YasuNeF1,AveMEAN5Kairi,AveMEAN25Kairi  FROM tkabuka where MeigaraCd = '%s' ORDER BY Hizuke DESC", ukabukaGroup.MeigaraCd)
		} else {
			wSQL = fmt.Sprintf("SELECT MeigaraCd,HizukeNum,Hizuke,HazimariNe,OwariNe,ZenOwariNe,HazimariNeF1,TakaNeF1,YasuNeF1,MEAN5Kairi,MEAN25Kairi  FROM tkabuka where MeigaraCd = '%s' and Hizuke >= '2020/5/1' ORDER BY Hizuke DESC", ukabukaGroup.MeigaraCd)
			wSQL = fmt.Sprintf("SELECT MeigaraCd,HizukeNum,Hizuke,HazimariNe,OwariNe,ZenOwariNe,HazimariNeF1,TakaNeF1,YasuNeF1,MEAN5Kairi,MEAN25Kairi  FROM tkabuka where MeigaraCd = '%s' ORDER BY Hizuke DESC", ukabukaGroup.MeigaraCd)
		}

		rows, err := db.Query(wSQL)
		if err != nil {
			panic(err.Error())
		}

		var wKabukasTmp VkabukaPool

		//一件目
		for rows.Next() {

			err := rows.Scan(&ukabuka.MeigaraCd, &ukabuka.HizukeNum, &ukabuka.Hizuke, &ukabuka.HazimariNe, &ukabuka.OwariNe, &ukabuka.ZenOwariNe, &ukabuka.HazimariNeF1, &ukabuka.TakaNeF1, &ukabuka.YasuNeF1, &ukabuka.MEAN5Kairi, &ukabuka.MEAN25Kairi)
			if err != nil {
				panic(err.Error())
			}
			defer rows.Close()
			wAns, wKabuka := com.ChkKabuka(ukabuka)

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
	sort.Sort(VkabukaPool(wKabukas))

	var wJyoukens []com.TJyouken
	var wHyokaMax float64
	var wKeyNo int

	if *aRunMode == "GEN" {
		wNo := 0
		if *aScenarioDBInit == false {
			wKeyNo, wHyokaMax = com.ScenarioDBNo(db, *aScenarioDB)
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
							wNo, wKeyNo, wJyoukens = setJyouken(wJyoukens, wNoSta, wNoEnd, wNo, wKeyNo, *aUriKai, k5Idx, k25Idx, waitIdx, nanIdx, uriIdx, SKoza,*aKozaOn,
								*aNanpinLvl,*aKaiSasiNe, *aNanpinOwariNe, *aNanpinYoriNe, *aUriOwariNe, *aUriYoriNe)
							// for hoyuIdx := 10; hoyuIdx <= 80; hoyuIdx += 10 {
							// 	for KozaMusiIdx := 0; KozaMusiIdx <= 0; KozaMusiIdx += 1 {
							// 		//for KozaMusiIdx := 0; KozaMusiIdx <= 1; KozaMusiIdx += 1 {
							// 		var wNanpinKozaMusiIdxMax int
							// 		var wNanpinMax int
							// 		if KozaMusiIdx == 0 {
							// 			wNanpinMax = 1
							// 			//wNanpinMax = 1
							// 			wNanpinKozaMusiIdxMax = 0
							// 			//wNanpinKozaMusiIdxMax = 1
							// 		}
							// 		for NanpinLvlIdx := 0; NanpinLvlIdx <= 2; NanpinLvlIdx += 1 {
							// 			for NanpinKozaMusiIdx := 0; NanpinKozaMusiIdx <= wNanpinKozaMusiIdxMax; NanpinKozaMusiIdx += 1 {
							// 				if NanpinKozaMusiIdx > 0 {
							// 					wNanpinMax = 0
							// 				}
							// 				for NanpinGenKinIdx := 0; NanpinGenKinIdx <= wNanpinMax; NanpinGenKinIdx += 1 {
							// 					if wNoSta <= wNo && wNo <= wNoEnd {
							// 						var wJyouken TJyouken
							// 						wJyouken.No = wKeyNo
							// 						wJyouken.KBN = *aUriKai
							// 						wJyouken.Kairi5 = float64(k5Idx)
							// 						wJyouken.Kairi25 = float64(k25Idx)
							// 						wJyouken.WaitDay = float64(waitIdx)
							// 						wJyouken.HoyuDay = float64(hoyuIdx)
							// 						wJyouken.Nanpin = float64(nanIdx) / 100
							// 						wJyouken.UriRitu = float64(uriIdx) / 100
							// 						wJyouken.SonekiSum = 0.0
							// 						wJyouken.SonekiAve = 0.0
							// 						wJyouken.StaKoza = SKoza
							// 						wJyouken.KozaMusi = KozaMusiIdx
							// 						wJyouken.NanpinLvl = NanpinLvlIdx
							// 						wJyouken.NanpinKozaMusi = NanpinKozaMusiIdx
							// 						wJyouken.NanpinGenKin = NanpinGenKinIdx
							// 						wJyoukens = append(wJyoukens, wJyouken)
							// 						wKeyNo++
							// 						//fmt.Println(wJyouken)
							// 					}
							// 					wNo++
							// 				}
							// 			}
							// 		}
							// 	}
							// }
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
							wNo, wKeyNo, wJyoukens = setJyouken(wJyoukens, wNoSta, wNoEnd, wNo, wKeyNo, *aUriKai, k5Idx, k25Idx, waitIdx, nanIdx, uriIdx, SKoza,*aKozaOn,
								*aNanpinLvl,*aKaiSasiNe, *aNanpinOwariNe, *aNanpinYoriNe, *aUriOwariNe, *aUriYoriNe)
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
			wVJyouken := com.GetVJyokenScan(rows)
			wAns, wJyouken := com.ChkJyouken(wVJyouken)
			if wAns {
				if wKairi5 != wVJyouken.Kairi5.Float64 || wKairi25 != wVJyouken.Kairi25.Float64 {
					// wSQL := fmt.Sprintf("INSERT INTO tJyouken%s VALUES ( %v,'%s',%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v)",
					// *aSimDB,
					// wJyouken.No,
					// wJyouken.KBN,
					// wJyouken.Kairi5,
					// wJyouken.Kairi25,
					// wJyouken.Nanpin,
					// wJyouken.UriRitu,
					// wJyouken.WaitDay,
					// wJyouken.StaKoza,
					// wJyouken.KozaMusi,
					// wJyouken.NanpinLvl,
					// wJyouken.NanpinKozaMusi,
					// wJyouken.NanpinGenKin,
					// wJyouken.HoyuDay,
					// wJyouken.Cnt,
					// wJyouken.SonekiSum,
					// wJyouken.SonekiAve,
					// wJyouken.HoyuDayAvg,
					// wJyouken.Hyoka,
					// wJyouken.CntHoyutyu,
					// wJyouken.CntHusoku,
					// wJyouken.CntCancel,
					// wJyouken.SyoRitu)
					// fmt.Println(wSQL)
					// rows1, err := db.Query(wSQL)
					// if err != nil {
					// 	panic(err.Error())
					// }
					// rows1.Close()
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
		com.InitDb(db, delTable, *aScenarioDB)
	} else if *aRunMode == "SIM" {
		com.InitDb(db, delTable, *aSimDB)
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
		go func(wJyouken com.TJyouken) {
			defer wg.Done()
			semaphore <- i
			//wJyoukenKekka, wSimbaibaikekkaList := runBuySim(SW, wJyouken, wKabukas)
			wJyoukenKekka, wSimbaibaikekkaList, wJyoukenLogs := runBuySim(*aScenarioDB, wJyouken, wKabukas)
			fmt.Println(wJyoukenKekka)
			if wJyoukenKekka.Hyoka > 100 || *aRunMode == "SIM" {
				com.InsertJyouken(db,OutDb,wJyoukenKekka)
				// wSQL := fmt.Sprintf("INSERT INTO tJyouken%s VALUES ( %v,'%s',%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v)",
				// OutDb,
				// 	wJyoukenKekka.ScenarioNo,
				// 	wJyoukenKekka.KBN,
				// 	wJyoukenKekka.Kairi5,
				// 	wJyoukenKekka.Kairi25,
				// 	wJyoukenKekka.KaiHoho,
				// 	wJyoukenKekka.Nanpin,
				// 	wJyoukenKekka.NanpinKaiHoho,
				// 	wJyoukenKekka.UriRitu,
				// 	wJyoukenKekka.UriHoho,
				// 	wJyoukenKekka.StaKoza,
				// 	wJyoukenKekka.KozaMusi,
				// 	wJyoukenKekka.NanpinLvl,
				// 	wJyoukenKekka.HoyuDay,
				// 	wJyoukenKekka.Cnt,
				// 	wJyoukenKekka.SonekiSum,
				// 	wJyoukenKekka.SonekiAve,
				// 	wJyoukenKekka.HoyuDayAvg,
				// 	wJyoukenKekka.Hyoka,
				// 	wJyoukenKekka.CntHoyutyu,
				// 	wJyoukenKekka.CntHusoku,
				// 	wJyoukenKekka.CntCancel,
				// 	wJyoukenKekka.SyoRitu)
				// rows1, err := db.Query(wSQL)
				// if err != nil {
				// 	panic(err.Error())
				// }
				// rows1.Close()

				if wHyokaMax < wJyoukenKekka.Hyoka || *aRunMode == "SIM" {
					wHyokaMax = wJyoukenKekka.Hyoka
					for j, wSimbaibaikekka := range wSimbaibaikekkaList {
						//if (wSimbaibaikekkaList[i].UriIdx >0){
						wSimbaibaikekka.LogIdx = j
						wSimbaibaikekka.ScenarioNo = wJyoukenKekka.ScenarioNo
						wSimbaibaikekka.Kbn = "00"
						com.InsertSimbaibaikekka(db,OutDb,wSimbaibaikekka)
						// wSQL2 := fmt.Sprintf("INSERT INTO tsimbaibaikekka%s VALUES ( NOW(),%v,'%s',%v,%v,'%s','%s',%v,%v,%v,%v,%v,%v,'%s',%v,%v,'%s',%v,%v,%v,'%s',%v,%v,'%s',%v,%v,%v,'%s',%v,%v,%v,'%s',%v,%v,'%s',%v,%v,%v,%v,%v,%v,%v,%v,%v,%v)",
						// OutDb,
						// 	wSimbaibaikekka.LogIdx,
						// 	wSimbaibaikekka.Hizuke,
						// 	wSimbaibaikekka.ScenarioNo,
						// 	wSimbaibaikekka.KaiIdx,
						// 	wSimbaibaikekka.MeigaraCd,
						// 	wSimbaibaikekka.Kbn,
						// 	wSimbaibaikekka.KaiKabugaku,
						// 	wSimbaibaikekka.KaiKabuka,
						// 	wSimbaibaikekka.KaiKabusu,
						// 	wSimbaibaikekka.KaiKabugaku1,
						// 	wSimbaibaikekka.KaiKabuka1,
						// 	wSimbaibaikekka.KaiKabusu1,
						// 	wSimbaibaikekka.KaiTiming1,
						// 	wSimbaibaikekka.KaiHusoku2,
						// 	wSimbaibaikekka.KaiIdx2,
						// 	wSimbaibaikekka.KaiHizuke2,
						// 	wSimbaibaikekka.KaiKabugaku2,
						// 	wSimbaibaikekka.KaiKabuka2,
						// 	wSimbaibaikekka.KaiKabusu2,
						// 	wSimbaibaikekka.KaiTiming2,
						// 	wSimbaibaikekka.KaiHusoku3,
						// 	wSimbaibaikekka.KaiIdx3,
						// 	wSimbaibaikekka.KaiHizuke3,
						// 	wSimbaibaikekka.KaiKabugaku3,
						// 	wSimbaibaikekka.KaiKabuka3,
						// 	wSimbaibaikekka.KaiKabusu3,
						// 	wSimbaibaikekka.KaiTiming3,
						// 	wSimbaibaikekka.NanpinYotei,
						// 	wSimbaibaikekka.UriYotei,
						// 	wSimbaibaikekka.UriIdx,
						// 	wSimbaibaikekka.UriHizuke,
						// 	wSimbaibaikekka.UriKabugaku,
						// 	wSimbaibaikekka.UriKabuka,
						// 	wSimbaibaikekka.UriTiming,
						// 	wSimbaibaikekka.TesuRyo,
						// 	wSimbaibaikekka.Kinri,
						// 	wSimbaibaikekka.Soneki,
						// 	wSimbaibaikekka.SonekiRitu,
						// 	wSimbaibaikekka.Day,
						// 	wSimbaibaikekka.DayAll,
						// 	wSimbaibaikekka.MaxKabuka,
						// 	wSimbaibaikekka.MaxKabukaRitu,
						// 	wSimbaibaikekka.MinKabuka,
						// 	wSimbaibaikekka.MinKabukaRitu)
						//fmt.Println(wSQL2)
						// rows2, err2 := db.Query(wSQL2)
						// if err2 != nil {
						// 	panic(err2.Error())
						// }
						// rows2.Close()
					}

					for _, wJyoukenLog := range wJyoukenLogs {
						com.InsertJyokenLog(db,OutDb,wJyoukenLog)
						// wSQL3 := fmt.Sprintf("INSERT INTO tJyoukenLog%s VALUES ( %v,'%s',%v,%v,%v,%v,%v,%v,%v,%v)", OutDb, wJyoukenLog.No, wJyoukenLog.Hizuke, wJyoukenLog.Koza, wJyoukenLog.KaiGaku, wJyoukenLog.KaiGakuGoukei, wJyoukenLog.HyokaGaku, wJyoukenLog.Soneki, wJyoukenLog.SonekiRitu, wJyoukenLog.HyokaGakuGokei, wJyoukenLog.HoyuCnt)

						// //fmt.Println(wSQL3)
						// rows3, err := db.Query(wSQL3)
						// if err != nil {
						// 	panic(err.Error())
						// }
						// rows3.Close()
					}
				}
			}
			<-semaphore
		}(wJyouken)

	}
	wg.Wait()
}
