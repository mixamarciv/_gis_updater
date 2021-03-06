package main

import (
	"database/sql"
	"encoding/json"
	"strconv"

	_ "github.com/nakagami/firebirdsql"

	s "strings"

	mf "github.com/mixamarciv/gofncstd3000"
)

var db *sql.DB

func Initdb(opt map[string]interface{}) {
	checkOptionsAndExit(opt, []string{"Database"})
	dbopt := opt["Database"].(string)
	var err error
	db, err = sql.Open("firebirdsql", "sysdba:masterkey@"+dbopt)
	LogPrintErrAndExit("ошибка подключения к базе данных "+dbopt, err)
	LogPrint("установлено подключение к БД: " + dbopt)
}

//обновляем параметр opt из записи huis_uk.options
func LoadUkData(fcomp string) map[string]interface{} {
	LogPrint("выбираем данные по УК(" + fcomp + ")")

	query := `SELECT options FROM huis_uk
			  WHERE fcomp = '` + fcomp + `'
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
		LogPrintAndExit("ОШИБКА в БД данные по УК(" + fcomp + ") не найдены!")
	}

	d := FromStrToJson(options)
	return d
}

func FromStrToJson(str string) map[string]interface{} {
	str = s.Trim(str, " \t\r\n")
	if str[0:1] != "{" && str[0:1] != "[" {
		str = "{" + str + "}"
	}

	if str[0:1] == "{" {
		var d map[string]interface{}
		err := json.Unmarshal([]byte(str), &d)
		LogPrintErrAndExit("ОШИБКА разбора JSON строки в объект: "+str, err)
		return d
	}
	if str[0:1] == "[" {
		var d []string
		err := json.Unmarshal([]byte(str), &d)
		LogPrintErrAndExit("ОШИБКА разбора JSON строки в массив: "+str, err)
		d2 := make(map[string]interface{}, 1)
		for key, val := range d {
			key_s := strconv.Itoa(key)
			d2[key_s] = val
		}
		return d2
	}
	return make(map[string]interface{}, 1)
}

func FromJsonToStr(v interface{}) string {
	j, err := json.Marshal(v)
	LogPrintErrAndExit("ОШИБКА JSON преобразования в строку", err)
	return string(j)
}

func LoadOptFromDataFile(file string) map[string]interface{} {
	LogPrint("загружаем параметры из файла " + file)
	str := ReadFileOrExit(file)
	d := FromStrToJson(str)
	return d
}

func AddOptions(to map[string]interface{}, from interface{}) {
	for k, val := range from.(map[string]interface{}) {
		to[k] = val
	}
}

func ReadFileOrExit(file string) string {
	str, err := mf.FileReadStr(file)
	LogPrintErrAndExit("ОШИБКА чтения файла: \n"+file+"\n\n", err)
	return str
}
