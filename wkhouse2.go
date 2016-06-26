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
		d := make(map[string]interface{}, 1) //основные характеристики дома

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
		d["HouseManagementType_Code"] = node.SelectNode("*", "HouseManagementType").S("*", "Code")
		d["HouseManagementType_GUID"] = node.SelectNode("*", "HouseManagementType").S("*", "GUID")

		new_houseoptions := FromJsonToStr(d)
		if new_houseoptions != houseoptions {
			wkHouse_dbupdatehouseOptions(fcomp, houseguid, new_houseoptions)
		}

		opt["houseOptions"] = d
	}

	{
		e := make(map[string]interface{}, 1) //список подьездов с данными по квартирам
		entrs := node.SelectNodes("*", "Entrance")
		for i := 0; i < len(entrs); i++ {
			entrnode := entrs[i]
			num := entrnode.S("*", "EntranceNum")

			entr := make(map[string]interface{}, 1)
			entr["EntranceNum"] = num
			entr["StoreysCount"] = entrnode.S("*", "StoreysCount")
			entr["EntranceGUID"] = entrnode.S("*", "EntranceGUID")
			entr["flats"] = wkHouse_loadResidentialPremises(entrnode)
			e[num] = entr
		}

		wkHouse_dbCheckHouseFlats(fcomp, house, e)
		opt["entrences"] = e

	}

	//LogPrint(Fmts("== %d ============================================", 0))
	//LogPrint(Fmts("%#v", opt["entrences"]))
	//LogPrint(Fmts("==/%d ============================================", 0))

	//LogPrint("ответ на запрос текущих данных по дому получен")
	return make(map[string]interface{}, 1)
}

//загружаем список по квартирам
func wkHouse_loadResidentialPremises(node *xmlx.Node) map[string]interface{} {
	f := make(map[string]interface{}, 1) //список по квартирам
	flats := node.SelectNodes("*", "ResidentialPremises")
	for i := 0; i < len(flats); i++ {
		flat := flats[i]

		num := flat.S("*", "PremisesNum")

		fi := make(map[string]interface{}, 1)
		fi["PremisesNum"] = num
		fi["No_RSO_GKN_EGRP_Registered"] = flat.S("*", "No_RSO_GKN_EGRP_Registered")
		fi["PremisesCharacteristic_Code"] = flat.SelectNode("*", "PremisesCharacteristic").S("*", "Code")
		fi["PremisesCharacteristic_GUID"] = flat.SelectNode("*", "PremisesCharacteristic").S("*", "GUID")
		fi["TotalArea"] = flat.S("*", "TotalArea")
		fi["GrossArea"] = flat.S("*", "GrossArea")
		fi["PremisesUniqueNumber"] = flat.S("*", "PremisesUniqueNumber")
		fi["PremisesGUID"] = flat.S("*", "PremisesGUID")

		f[num] = fi
	}
	/*************************************
	<ns5:ResidentialPremises>
	    <ns5:No_RSO_GKN_EGRP_Registered>true</ns5:No_RSO_GKN_EGRP_Registered>
	    <ns5:PremisesNum>1</ns5:PremisesNum>
	    <ns5:EntranceNum>1</ns5:EntranceNum>
	    <ns5:PremisesCharacteristic>
	        <ns4:Code>1</ns4:Code>
	        <ns4:GUID>96a1ce61-b402-46c4-ac6e-34b8670480af</ns4:GUID>
	    </ns5:PremisesCharacteristic>
	    <ns5:TotalArea>40.0</ns5:TotalArea>
	    <ns5:GrossArea>30.0</ns5:GrossArea>
	    <ns5:PremisesUniqueNumber>8c51fa1e-a77d-4d01-bba1-9a4cf7e78266</ns5:PremisesUniqueNumber>
	    <ns5:ModificationDate>2016-03-28T17:19:34.271+03:00</ns5:ModificationDate>
	    <ns5:PremisesGUID>8c51fa1e-a77d-4d01-bba1-9a4cf7e78266</ns5:PremisesGUID>
	</ns5:ResidentialPremises>
	****************************************/
	return f
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

	//LogPrint(Fmts("== %d ============================================", 0))
	//LogPrint(Fmts("%#v", info))
	//LogPrint(Fmts("==/%d ============================================", 0))
	return
}

//загружаем и проверяем данные по квартирам в бд
func wkHouse_dbCheckHouseFlats(fcomp, house, e map[string]interface{}) {
	LogPrint("загружаем из БД данные по крвартирам")

	ne := make(map[string]interface{}, 1) //новый список подьездов с данными по квартирам

	exclude_list := ""
	for entrNum, entr := range e {
		flats := entr["flats"]
		for flatNum, flat := range flats {
			entrflatnum := entrNum + "-" + flatNum
			exclude_list = exclude_list + "'" + entrflatnum + "',"
		}
	}
	exclude_list = exclude_list[0 : len(exclude_list)-1]

	query := `SELECT  
	 			COALESCE(e.entrance,""),
				COALESCE(f.flat,""),
				COALESCE(e.options,"") AS entrance_options,
				COALESCE(f.options,"") AS flat_options,
				(SELECT COUNT(*) FROM t_obj_flat tf 
				 WHERE tf.strcode=h.strcode 
				   AND tf.house=h.house
				   AND tf.entrance=e.entrance) AS StoreysCount,
				COALESCE((SELECT MAX(k.ob_area) FROM kv2_kart k 
				 WHERE k.fcomp=h.fcomp 
				   AND k.fperiod=tt.fperiod 
				   AND k.strcode=h.strcode
				   AND k.house2=h.house
				   AND k.flat2=h.flat
				),20.0) AS ob_area,
				COALESCE((SELECT MAX(k.jil_area) FROM kv2_kart k 
				 WHERE k.fcomp=h.fcomp 
				   AND k.fperiod=tt.fperiod 
				   AND k.strcode=h.strcode
				   AND k.house2=h.house
				   AND k.flat2=h.flat
				),20.0) AS jil_area
		      FROM t_obj_house h
				 LEFT JOIN t_obj_entrance e ON e.strcode=h.strcode AND e.house=h.house
				 LEFT JOIN t_obj_flat f ON f.strcode=h.strcode AND f.entrance=e.entrance
				 LEFT JOIN t_kv2_uk_last_period tt ON tt.fcomp=h.fcomp
	          WHERE h.fiasguid LIKE '` + house + `' AND h.fcomp=` + fcomp + `
			    AND e.entrance||'-'||f.flat NOT IN (` + exclude_list + `)
			 `
	rows, err := db.Query(query)
	LogPrintErrAndExit("ОШИБКА выполнения запроса: \n"+query+"\n\n", err)

	var entrance, flat, entrance_options, flat_options, StoreysCount, ob_area, jil_area string
	found := 0
	for rows.Next() {
		found++
		err = rows.Scan(&entrance, &flat, &entrance_options, &flat_options, &StoreysCount, &ob_area, &jil_area)
		LogPrintErrAndExit("rows.Scan error: \n"+query+"\n\n", err)
	}
	if found == 0 {
		LogPrint("ВНИМАНИЕ: по заданным параметрам квартиры в базеданных не найдены!")
	}

	//LogPrintErrAndExit("ОШИБКА выполнения запроса: \n"+query+"\n\n", err)

	LogPrint(Fmts("== %d ============================================", 0))
	LogPrint(Fmts("%#v", info))
	LogPrint(Fmts("==/%d ============================================", 0))
	return
}
