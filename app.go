package main

import (
	"fmt"
	"os"

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
		House      string `long:"house" description:"FIAS house GUID"`
	}
	_, err := flags.ParseArgs(&opts, os.Args)
	LogPrintErrAndExit("ошибка разбора параметров", err)

	//LogPrint(Fmts("args: %#v", args))
	//LogPrint(Fmts("opts: %+v", opts))
	//LogPrint(Fmts("opts.Type: %#v", opts.Type))

	options := structs.Map(opts)

	Initdb(options)

	switch opts.Type {
	case `house`:
		S0300_house(options)
		return
	}

	LogPrint(Fmts("ОШИБКА: обработчик для параметра --type %+v не задан", opts.Type))
}

func checkOptionsAndExit(options map[string]interface{}, need []string) {
	var notfound []string = nil
	var strnotfound string = ""
	for _, param := range need {
		if options[param] == nil {
			notfound = append(notfound, param)
			strnotfound = strnotfound + " " + param
		}
	}
	if len(notfound) > 0 {
		LogPrint("ОШИБКА: не заданы параметры: " + strnotfound)
		LogPrint(Fmts("список указанных параметров: %+v", options))
		os.Exit(1)
	}
}
