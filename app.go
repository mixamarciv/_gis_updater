package main

import (
	"fmt"
	"os"
	"strings"

	mf "github.com/mixamarciv/gofncstd3000"

	flags "github.com/jessevdk/go-flags"

	structs "github.com/fatih/structs"
)

var Fmts = fmt.Sprintf

func init() {
	fmt.Printf(mf.CurTimeStrShort()[0:0] + "\n")
}

func main() {
	var opts struct {
		Database   string `long:"database" description:"ip:databasepath"`
		Cryptohost string `long:"cryptohost" description:"crypto tunnel host"`
		Asyncserv  string `long:"asyncserv" description:"async http service host"`
		Signserv   string `long:"signserv" description:"sign http service host"`
		Huisver    string `long:"huisver" description:"ver gis xml files"`
		Type       string `long:"type" description:"gis xml request type(house)"`
		Fcomp      string `long:"fcomp" description:"uk num"`
		Uktype     string `long:"uktype" description:"RSO or UO"`
		House      string `long:"house" description:"FIAS house GUID"`
	}
	_, err := flags.ParseArgs(&opts, os.Args)
	LogPrintErrAndExit("ошибка разбора параметров", err)

	//LogPrint(Fmts("args: %#v", args))
	//LogPrint(Fmts("opts: %+v", opts))
	//LogPrint(Fmts("opts.Type: %#v", opts.Type))

	options := structs.Map(opts)

	//создаем все параметры в нижнем регистре
	for key, val := range options {
		lkey := strings.ToLower(key)
		if key != strings.ToLower(key) {
			options[lkey] = val
		}
	}

	checkOptionsAndExit(options, []string{"Database"})
	Initdb(options)

	switch opts.Type {
	case `house`:
		wkHouse(options)
		return
	}

	LogPrint(Fmts("ОШИБКА: обработчик для параметра --type %+v не задан", opts.Type))
}

func checkOptionsAndExit(options map[string]interface{}, need []string) {
	var notfound []string = nil
	var strnotfound string = ""
	for _, param := range need {
		if val, ok := options[param]; !ok || val == nil || val == "" {
			notfound = append(notfound, param)
			strnotfound = strnotfound + " " + param
		}
	}
	if len(notfound) > 0 {
		LogPrint("ОШИБКА: не заданы обязательные параметры: " + strnotfound)
		LogPrint(Fmts("список указанных параметров: %+v", options))
		os.Exit(1)
	}
}
