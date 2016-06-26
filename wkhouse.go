package main

//s "strings"
//import xml "github.com/jteeuwen/go-pkg-xmlx"
//import xml "github.com/jteeuwen/go-pkg-xmlx"

import (
	"bytes"
	"text/template"

	mf "github.com/mixamarciv/gofncstd3000"
)

func wkHouse(opt map[string]interface{}) {
	checkOptionsAndExit(opt, []string{"House", "Fcomp", "Asyncserv", "Signserv", "Cryptohost", "Huisver", "Uktype"})
	LogPrint(Fmts("%+v", opt))

	fcomp := opt["Fcomp"].(string)
	house := opt["House"].(string)

	path, _ := mf.AppPath()
	path = path + "/templates/xml/" + opt["Huisver"].(string) + "/house"
	opt["templates_path"] = path
	opt["data_uk"] = LoadUkData(fcomp)
	AddOptions(opt, opt["data_uk"])

	opt["data_expopt"] = LoadOptFromDataFile(path + "/export.xml.json")
	opt["data_impopt"] = LoadOptFromDataFile(path + "/import_" + opt["Uktype"].(string) + ".xml.json")

	list := wkHouse_loadhouselist(fcomp, house)
	for _, h := range list {
		wkHouse_work(opt, h)
	}
}

//загружаем список домов которые попадают под указанные критерии
func wkHouse_loadhouselist(fcomp, house string) []string {
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
func wkHouse_work(opt map[string]interface{}, house string) {
	opt["FIASHouseGuid"] = house
	wkHouse_work_1_getcurdata(opt, house)
	LogPrint(Fmts("%+v", opt))
}

//запрашиваем текущие данные у гиса
func wkHouse_work_1_getcurdata(opt map[string]interface{}, house string) string /* *xml.Node */ {
	xml := wkHouse_render_exportxml(opt, house)
	LogPrint(Fmts("xml:%+v", xml))
	return house
}

func wkHouse_render_exportxml(opt map[string]interface{}, house string) string {
	AddOptions(opt, opt["data_expopt"])

	file := opt["templates_path"].(string) + "/export.xml"
	file_str, err := mf.FileReadStr(file)
	LogPrintErrAndExit("ОШИБКА чтения файла: \n"+file+"\n\n", err)

	funcMap := template.FuncMap{
		"RandomGUID":   mf.StrUuid,
		"CurDateTime1": mf.CurTimeStrRFC3339,
		"CurDateTime2": mf.CurTimeStr,
	}

	type UserVars struct {
		CurDateTime string
		HuisVer     string
		Data        map[string]interface{}
	}

	vars := new(UserVars)
	vars.CurDateTime = mf.CurTimeStr()
	vars.RandomGUID1 = mf.Uuid()
	vars.HuisVer = opt["Huisver"].(string)
	vars.Data = opt

	t1, err := template.New("xml").Funcs(funcMap).Parse(file_str)
	LogPrintErrAndExit("parse template file error: \n"+file+"\n\n", err)

	buff1 := new(bytes.Buffer)
	err = t1.Execute(buff1, vars)
	LogPrintErrAndExit("render template file error: \n"+file+"\n\n", err)
	return string(buff1.Bytes())
}
