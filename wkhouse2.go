package main

//s "strings"
//import xml "github.com/jteeuwen/go-pkg-xmlx"
//import xml "github.com/jteeuwen/go-pkg-xmlx"

import (
	xmlx "github.com/jteeuwen/go-pkg-xmlx"
	//mf "github.com/mixamarciv/gofncstd3000"
)

//проверяем какие данные уже загружены, если все ок ставим check_date = mf.CurTimeStrRFC3339
func wkHouse_work_2_getNeedUpdateData(opt map[string]interface{}, houseguid string, node *xmlx.Node) map[string]interface{} {
	LogPrint("загружаем список данных необходимых для обновления")
	fcomp := opt["fcomp"].(string)
	{
		d := make(map[string]interface{}, 1)

		houseoptions := wkHouse_dbloadhouseOptions(fcomp, houseguid)

		node = node.SelectNode("*", "exportHouseResult")

		d["HouseUniqueNumber"] = node.S("*", "HouseUniqueNumber")
		d["UndergroundFloorCount"] = node.S("*", "UndergroundFloorCount")
		d["MinFloorCount"] = node.S("*", "MinFloorCount")

		nodebase := node.SelectNode("*", "BasicCharacteristicts")
		d["No_RSO_GKN_EGRP_Registered"] = nodebase.S("*", "No_RSO_GKN_EGRP_Registered")
		d["TotalSquare"] = nodebase.S("*", "TotalSquare")
		d["State_Code"] = nodebase.SelectNode("*", "State").S("*", "Code")
		d["State_GUID"] = nodebase.SelectNode("*", "State").S("*", "GUID")
		d["UsedYear"] = nodebase.S("*", "UsedYear")
		d["FloorCount"] = nodebase.S("*", "FloorCount")
		d["OlsonTZ_Code"] = nodebase.SelectNode("*", "OlsonTZ").S("*", "Code")
		d["OlsonTZ_GUID"] = nodebase.SelectNode("*", "OlsonTZ").S("*", "GUID")
		d["CulturalHeritage"] = nodebase.S("*", "CulturalHeritage")

		new_houseoptions := FromJsonToStr(d)
		if new_houseoptions != houseoptions {
			wkHouse_dbupdatehouseOptions(fcomp, houseguid, new_houseoptions)
		}
	}

	//LogPrint(Fmts("== %d ============================================", 0))
	//LogPrint(Fmts("%#v", houseoptions))
	//LogPrint(Fmts("==/%d ============================================", 0))

	//LogPrint("ответ на запрос текущих данных по дому получен")
	return make(map[string]interface{}, 1)
}

//загружаем данные поля options у выбранного дома
func wkHouse_dbloadhouseOptions(fcomp, house string) string {
	LogPrint("загружаем данные поля options у дома " + house)
	query := `SELECT COALESCE(h.options,'') FROM t_obj_house h
	          WHERE h.fiasguid LIKE '` + house + `' AND fcomp=` + fcomp + `
			 `
	rows, err := db.Query(query)
	LogPrintErrAndExit("ОШИБКА выполнения запроса: \n"+query+"\n\n", err)

	var options string
	found := 0
	for rows.Next() {
		found++
		err = rows.Scan(&options)
		LogPrintErrAndExit("rows.Scan error: \n"+query+"\n\n", err)
	}
	if found == 0 {
		LogPrintAndExit("по заданным параметрам дома в базеданных не найдены!")
	}
	return options
}

//загружаем данные поля options у выбранного дома
func wkHouse_dbupdatehouseOptions(fcomp, house, options string) {
	LogPrint("обновляем поля options у дома")
	query := `UPDATE t_obj_house h SET h.options=?
	          WHERE h.fiasguid LIKE '` + house + `' AND fcomp=` + fcomp + ` 
			 `
	info, err := db.Exec(query, options)

	LogPrintErrAndExit("ОШИБКА выполнения запроса: \n"+query+"\n\n", err)

	LogPrint(Fmts("== %d ============================================", 0))
	LogPrint(Fmts("%#v", info))
	LogPrint(Fmts("==/%d ============================================", 0))
	return
}
