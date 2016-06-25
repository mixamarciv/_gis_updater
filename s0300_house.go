package main

//s "strings"
//import xml "github.com/jteeuwen/go-pkg-xmlx"
//import xml "github.com/jteeuwen/go-pkg-xmlx"

import mf "github.com/mixamarciv/gofncstd3000"

func S0300_house(opt map[string]interface{}) {
	checkOptionsAndExit(opt, []string{"House", "Fcomp", "Asyncserv", "Signserv", "Cryptohost", "Huisver"})

	fcomp := opt["Fcomp"].(string)
	house := opt["House"].(string)

	loadUkData(opt, fcomp)

	list := loadhouselist(fcomp, house)

	for _, h := range list {
		work(opt, h)
	}
}

//загружаем список домов которые попадают под указанные критерии
func loadhouselist(fcomp, house string) []string {
	LogPrint("выбираем список домов по указанным критериям fcomp:" + fcomp + " house:" + house)
	var ret []string
	query := `SELECT h.fiasguid,s.street||' '||h.house AS info FROM t_obj_house h
				LEFT JOIN street_kladr s ON s.strcode=h.strcode
	          WHERE h.fiasguid LIKE '` + house + `'
				AND h.fcomp = '` + fcomp + `'
			`
	rows, err := db.Query(query)
	LogPrintErrAndExit("ОШИБКА выполнения запроса: \n"+query+"\n\n", err)

	var fiasguid, info string
	found := 0
	for rows.Next() {
		found++
		err = rows.Scan(&fiasguid, &info)
		LogPrintErrAndExit("rows.Scan error: \n"+query+"\n\n", err)

		LogPrint(fiasguid + "| " + info)
		ret = append(ret, fiasguid)

	}
	if found == 0 {
		LogPrintAndExit("по заданным параметрам дома в базеданных не найдены!")
	}
	return ret
}

//работаем с отдельным домиком:
// 1 - запрашиваем у гиса данные на текущий момент
// 2 - сравниваем с тем что у нас в базе и сохраняем доп информацию
// 3 - если есть недостающие данные загружаем их в гис
// 4 - переходим к пункту 1
func work(opt map[string]interface{}, house string) {
	d1 := work_1_getcurdata(opt, house)
	LogPrint(d1)
}

func work_1_getcurdata(opt map[string]interface{}, house string) string /* *xml.Node */ {
	opt["FIASHouseGuid"] = house
	s, _ := mf.AppPath()
	return s
}
