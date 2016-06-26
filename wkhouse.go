package main

//s "strings"
//import xml "github.com/jteeuwen/go-pkg-xmlx"
//import xml "github.com/jteeuwen/go-pkg-xmlx"

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
	"time"

	xmlx "github.com/jteeuwen/go-pkg-xmlx"
	mf "github.com/mixamarciv/gofncstd3000"
	"github.com/parnurzeal/gorequest"
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
// 2 - сравниваем результат с тем что у нас в базе, и сохраняем доп информацию по дому в базу
// 4 - если есть недостающие данные загружаем их в гис
// 5 - переходим к пункту 1
func wkHouse_work(opt map[string]interface{}, house string) {
	LogPrint("start work with: " + house)
	opt["FIASHouseGuid"] = house

	node := wkHouse_work_1_getcurdata(opt, house)
	node = node.SelectNode("*", "exportHouseResult")

	needUpdateData := wkHouse_work_2_getNeedUpdateData(opt, house, node)

	LogPrint(Fmts("== %d ============================================", 2))
	LogPrint(Fmts("%#v", needUpdateData))
	LogPrint(Fmts("==/%d ============================================", 2))

	LogPrint("end work with: " + house)
}

//запрашиваем текущие данные у гиса
func wkHouse_work_1_getcurdata(opt map[string]interface{}, house string) *xmlx.Node {
	LogPrint("отправляем запрос текущих данных по дому")
	xml := wkHouse_render_exportxml(opt, house)

	//LogPrint(Fmts("%#v", opt))
	opt["url"] = opt["cryptohost"].(string) + opt["url"].(string)
	ret := make(map[string]string)
	ret["xml"] = xml
	ret["data"] = FromJsonToStr(opt)

	json_str := FromJsonToStr(ret)

	url := opt["asyncserv"].(string)
	result := sendRequestAsyncServ(url, string(json_str))

	resjson := FromStrToJson(result)
	resxml := resjson["0"].(string)

	doc := xmlx.New()
	err := doc.LoadString(resxml, nil)
	LogPrintErrAndExit("xmlx.LoadString error: \n"+resxml+"\n\n", err)
	node := doc.SelectNode("*", "Body")

	//LogPrint("ответ на запрос текущих данных по дому получен")
	return node
}

func wkHouse_render_exportxml(opt map[string]interface{}, house string) string {
	AddOptions(opt, opt["data_expopt"])

	file := opt["templates_path"].(string) + "/export.xml"
	file_str := ReadFileOrExit(file)
	xml := RenderTemplate(opt, file_str, file)
	return xml
}

func sendRequestSyncServ(url, body string) string {
	LogPrint("отправляем sync запрос: " + url)
	req := gorequest.New().Post(url)
	_, bodyresult, errs := req.Send(body).End()
	if len(errs) > 0 {
		LogPrint(Fmts("%#v", errs))
		LogPrintAndExit("request send error: \n url: " + url + "\n\n")
	}
	return bodyresult
}

func sendRequestAsyncServ(url, body string) string {
	LogPrint("отправляем async запрос: " + url)
	req := gorequest.New().Post(url)
	_, bodyresult, errs := req.Send(body).End()
	if len(errs) > 0 {
		LogPrint(Fmts("%#v", errs))
		LogPrintAndExit("request send error: \n url: " + url + "\n\n")
	}
	fmt.Print(".")

	result := make(chan string)
	go func() {
		for i := 1; i <= 1000; i++ {
			time.Sleep(time.Second * 1)
			fmt.Print(".")

			req := gorequest.New().Post(url)
			_, bodyresult, errs := req.Send(body).End()
			if len(errs) > 0 {
				LogPrint(Fmts("%#v", errs))
				LogPrintAndExit("request send error: \n url: " + url + "\n\n")
			}

			j := FromStrToJson(bodyresult)

			isok := checkIsEndResult(j)
			if isok > 0 {
				result <- bodyresult
				return
			}
		}
		LogPrintAndExit("Async request send error: timeout or bad result")
		result <- "the end"
	}()

	bodyresult = <-result
	fmt.Print("\n")
	return bodyresult
}

func checkIsEndResult(j map[string]interface{}) int {
	xml := j["0"].(string)
	i := strings.Index(xml, "RequestState>3</")
	if i > 0 {
		return 1
	}
	return 0
}

func RenderTemplate(opt map[string]interface{}, template_str string, debuginf string) string {
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
	vars.HuisVer = opt["Huisver"].(string)
	vars.Data = opt

	t1, err := template.New("xml").Funcs(funcMap).Parse(template_str)
	LogPrintErrAndExit("parse template error: \n"+debuginf+"\n\n", err)

	buff1 := new(bytes.Buffer)
	err = t1.Execute(buff1, vars)
	LogPrintErrAndExit("render template error: \n"+debuginf+"\n\n", err)
	return string(buff1.Bytes())
}
