::получаем curpath:
@FOR /f %%i IN ("%0") DO SET curpath=%~dp0
::задаем основные переменные окружения
@CALL "%curpath%/set_path.bat"


@del app.exe
@CLS

@echo === build =====================================================================
go build -o app.exe

@echo ==== start ====================================================================
@SET database=192.168.1.10/d:/_db_web/db002/0002.fdb
@SET cryptohost=http://192.168.1.163:8080
@SET asyncserv=http://192.168.1.120:8001/asyncreq
@SET signserv=http://192.168.1.23:8001/sign
::@SET signname=guk0001
@SET huisver=8.7.2.2
@SET type=house
@SET fcomp=46
@SET house=3db255a7-1325-4416-8c31-946aa3b150ef


@SET opt=--database %database% 
@SET opt=%opt% --cryptohost %cryptohost% --asyncserv %asyncserv% --signserv %signserv% 
@SET opt=%opt% --huisver %huisver% --type %type% --fcomp %fcomp% --house %house%


app.exe %opt%

@echo ==== end ======================================================================
@PAUSE
