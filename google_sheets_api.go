package main

import (
	"io/ioutil"
	"gopkg.in/Iwark/spreadsheet.v2"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"log"
	"regexp"
	"strconv"
)

var service = new(spreadsheet.Service)
var sprsheet = *new(spreadsheet.Spreadsheet)

func initialize_sheet() {
	data, err := ioutil.ReadFile(config.PathToGoogleKeyJson)
	if err != nil {
		log.Fatal("Cannot read from json key file: ", config.PathToGoogleKeyJson, ";")
	}
	conf, err := google.JWTConfigFromJSON(data, spreadsheet.Scope)
	checkError(err)
	client := conf.Client(context.TODO())
	service = spreadsheet.NewServiceWithClient(client)
}

func updSheet(id string) (error) {
	err := *new(error)
	sprsheet, err = service.FetchSpreadsheet(id)
	return err
}

func getPhoneList(userTableTitle string) ([]User) {
	var output = []User{}
	log.Println("getting phone list...")
	sheet, err := sprsheet.SheetByTitle(userTableTitle)
	checkError(err)
	err = sheet.Synchronize()
	checkError(err)
	if !checkAdminsHeader(sheet) {
		sheet.Update(0, 0, "name")
		sheet.Update(0, 1, "phone")
		sheet.Update(0, 2, "email")
	}
	lastRowId := getEmptyRow(sheet)
	for i := 1; i < lastRowId; i++ {
		row := sheet.Rows[i]
		output = append(output, User{Name: row[0].Value, Phone: row[1].Value, Email: row[2].Value})
	}
	log.Println("output of get phone list: ", output)
	return output
}

func getLastEntry(parameterList []string, wnumber string, serviceId string, userTableTitle string, tableTitle string)([]string) {
	sheet,err := new(spreadsheet.Sheet),*new(error)
	if tableTitle == "" {
		sheet, err = sprsheet.SheetByIndex(0)
	}else{
		sheet, err = sprsheet.SheetByTitle(tableTitle)
		if err != nil {
			sheet, err = sprsheet.SheetByIndex(0)
		}
	}
	checkErrorWithPush(err,serviceId,userTableTitle)
	err = sheet.Synchronize()
	checkErrorWithPush(err, serviceId, userTableTitle)
	out := []string{}
	lastId, ok := getLastValidRowId(sheet, wnumber)
	if ok {
		latestRow := sheet.Rows[lastId:lastId+1][0]
		log.Println("latestRow: ", latestRow, "; Params: ", parameterList)
		for _, parameter := range parameterList {
			appendCount := 0
			if parameter != "wnumber" { //do not return wnumber
				for i, v := range latestRow {
					if sheet.Rows[0][i].Value == parameter {
						out = append(out, v.Value)
						appendCount++
					}
				}
				if appendCount == 0 {
					out = append(out, "None/Column not exist")
				}
			} else {
				out = append(out, "wnumber")
			}
		}
		return out
	} else {
		return nil
	}
}

func getParameterNamesInOrder(order []string, locale string, translationTableTitle string) ([]string) {
	var out = []string{}
	sheet, err := sprsheet.SheetByTitle(translationTableTitle)
	if err != nil {
		return order // If no such sheet return just order
	}
	err = sheet.Synchronize()
	checkError(err)
	checkTranslationTableStructure(sheet)
	for i, parameterName := range order {
		log.Println("Parameter name: ", parameterName)
		localedParameterName := getParameterName(sheet, parameterName, locale)
		if localedParameterName == "" {
			localedParameterName = order[i]
		}
		log.Println("Value: " + localedParameterName)
		out = append(out, localedParameterName)
	}
	err = sheet.Synchronize()
	checkError(err)
	log.Println("Done!")
	return out
}

func addEntry(timestamp string, user_id string, protocol string, wnumber string, markable bool, mark int, params map[string]string, serviceId string, userTableTitle string) {
	log.Print("Adding entry: ", timestamp, " ", user_id, " ", protocol, " ", wnumber, " ", params)
	sheet, err := sprsheet.SheetByIndex(0)
	checkErrorWithPush(err, serviceId, userTableTitle)
	err = sheet.Synchronize()
	checkErrorWithPush(err, serviceId, userTableTitle)
	if err != nil {
		return
	}
	log.Print("Synchronized sheet...")

	//log.Println("sheet rows [0]=",sheet.Rows[0])
	if !checkHeader(sheet) {
		sheet.Update(0, 0, "timestamp")
		sheet.Update(0, 1, "user_id")
		//sheet.Update(0,2,"telegram_id")
		sheet.Update(0, 2, "protocol")
		sheet.Update(0, 3, "wnumber")
		sheet.Update(0, 4, "mark")
		//err = sheet.Synchronize()
		//checkError(err)
		//go addEntry(timestamp,user_id,protocol,wnumber,markable,mark,params)
		//return
	}
	//log.Println("rows: ",sheet.Rows)
	err = sheet.Synchronize()
	checkError(err)
	pgNamesCells := sheet.Rows[0][5:]
	pgNames := []string{}
	for _, cell := range pgNamesCells {
		pgNames = append(pgNames, cell.Value)
	}

	emptyRowIdx := getEmptyRow(sheet)
	emptyColumnIdx := getEmptyColumn(sheet)
	log.Println("Table shape: ", emptyColumnIdx, emptyRowIdx)
	//fill mark column
	fillWholeSheet(sheet, emptyRowIdx, emptyColumnIdx)
	emptyRowIdx = getEmptyRow(sheet)
	emptyColumnIdx = getEmptyColumn(sheet)

	log.Print(".")
	r := regexp.MustCompile(`^[a-z0-9]{8}(-[a-z0-9]{4}){3}-[a-z0-9]{12}$`)
	r1 := regexp.MustCompile(`^page([0-9]+)`)
	match := !r.MatchString(user_id) // if user id not wnumber
	if match {
		sheet.Update(emptyRowIdx, 0, timestamp)
		sheet.Update(emptyRowIdx, 1, user_id)
		sheet.Update(emptyRowIdx, 2, protocol)
		//sheet.Update(emptyRowIdx, 2, tgId)
		sheet.Update(emptyRowIdx, 3, wnumber)
	} else {
		sheet.Update(emptyRowIdx, 0, timestamp)
		sheet.Update(emptyRowIdx, 1, "0")
		sheet.Update(emptyRowIdx, 2, protocol)
		//sheet.Update(emptyRowIdx, 2, tgId)
		sheet.Update(emptyRowIdx, 3, user_id)
	}
	//123
	if markable {
		sheet.Update(emptyRowIdx, 4, strconv.Itoa(mark))
	} else {
		sheet.Update(emptyRowIdx, 4, "not markable")
	}

	for key, value := range params {
		//log.Println("Cols: ",sheet.Columns[0],"; rows: ",sheet.Rows[0])
		emptyRowIdx = getEmptyRow(sheet) - 1
		emptyColumnIdx = getEmptyColumn(sheet)
		//log.Println("iterating params: key:",key,"=",value,";")
		if !r1.MatchString(key) { //check if the pg name is not default (like page11)
			if !contains(pgNames, key) {
				//log.Println("!Contains; len(sheet.Rows[0])=",emptyColumnIdx)
				//fillColumn(sheet, emptyColumnIdx, emptyRowIdx)
				//fillUnfilledCols(sheet, emptyRowIdx, emptyColumnIdx)
				sheet.Update(0, emptyColumnIdx, key)
				//log.Println("Setting value")
				sheet.Update(emptyRowIdx, emptyColumnIdx, value)
				//log.Println("ended updating...")
			} else {
				//log.Println("Filling: ",emptyRowIdx, findColumn(sheet, key), value)
				//fillUnfilledCols(sheet, emptyRowIdx, emptyColumnIdx)
				sheet.Update(emptyRowIdx, findColumn(sheet, key), value) // update sheet
			}
		}
	}
	log.Print(".")
	err = sheet.Synchronize()
	checkError(err)
	log.Println("Done!")
}

func getParameterName(sheet *spreadsheet.Sheet, parameterName string, locale string) (string) {
	var out = ""
	for i, value := range sheet.Columns[0] {
		if value.Value == parameterName {
			out = sheet.Columns[findColumnNoPadding(sheet, locale)][i].Value
		}
	}
	return out
}

func checkHeader(sheet *spreadsheet.Sheet) (bool) {
	if len(sheet.Rows[0]) < 5 {
		return false
	} else if sheet.Rows[0][0].Value != "timestamp" || sheet.Rows[0][1].Value != "user_id" || sheet.Rows[0][2].Value != "protocol" || sheet.Rows[0][3].Value != "wnumber" || sheet.Rows[0][4].Value != "mark" {
		return false
	}
	return true
}

func checkAdminsHeader(sheet *spreadsheet.Sheet) (bool) {
	if len(sheet.Rows[0]) < 3 {
		return false
	} else if sheet.Rows[0][0].Value != "name" || sheet.Rows[0][1].Value != "phone" || sheet.Rows[0][2].Value != "mail" {
		return false
	}
	return true
}

func checkTranslationTableStructure(sheet *spreadsheet.Sheet) {
	log.Println("Checking translation table structure...")
	if sheet.Rows[0][0].Value != "parameters \\ language codes" {
		log.Println("Structure wrong!")
		mainSheet, err := sprsheet.SheetByIndex(0)
		checkError(err)
		err = mainSheet.Synchronize()
		checkError(err)
		sheet.Update(0, 0, "parameters \\ language codes")
		sheet.Update(1, 0, "prefix")
		sheet.Update(2, 0, "pushPrefix")
		sheet.Update(3, 0, "mailPostfix")
		sheet.Update(4, 0, "mailSubject")
		i := 5
		for _, val := range mainSheet.Rows[0] {
			log.Println("Updating: ", i, 0, val.Value)
			sheet.Update(i, 0, val.Value)
			i += 1
		}
	}
}

func getEmptyColumn(sheet *spreadsheet.Sheet) (int) {
	cells := sheet.Rows[0]
	out := len(cells)
	for i, cell := range cells {
		//log.Println("cell[",i,"]=",cell.Value)
		if cell.Value == "" {
			//log.Println("Found emty cell!(",i,")")
			out = i
			break
		}
	}
	return out
}

func getEmptyRow(sheet *spreadsheet.Sheet) (int) {
	//lengths := []int{}
	//appended := false
	//for j := 0; j <= getEmptyColumn(sheet); j++ {
	cells := sheet.Columns[0]
	out := len(cells)
	for i, cell := range cells {
		//log.Println("cell[", i, "]=", cell.Value)
		if cell.Value == "" {
			//log.Println("Found emty cell!(", i, ")")
			//lengths = append(lengths, i)
			//appended = true
			out = i
			break
		}
	}
	//if !appended {
	//	lengths = append(lengths, len(cells))
	//	appended = !appended
	//}
	//}
	//log.Println("lengths: ",lengths, "; Max: ",max(lengths))
	return out //max(lengths)
}

func getLastValidRowId(sheet *spreadsheet.Sheet, wnumber string) (int, bool) {
	lastEmptyRow := getEmptyRow(sheet) - 1
	prevI := lastEmptyRow
	log.Println("Last empty row: ", lastEmptyRow, wnumber)
	wnumberColumnId := findColumnNoPadding(sheet, "wnumber")
	if wnumberColumnId == 0{
		wnumberColumnId=3 //default values
	}
	userIdColumnId := findColumnNoPadding(sheet, "user_id")
	if userIdColumnId == 0{
		userIdColumnId = 1 //default values
	}
	log.Println("Wnumber column id: ",wnumberColumnId, "; uid cid: ",userIdColumnId)
	for i := lastEmptyRow; i >= 1; i -= 1 {
		log.Println("I:", i, sheet.Rows[i])
		log.Println("sheet.Rows[i][3].Value", sheet.Rows[i][wnumberColumnId].Value)
		log.Println("sheet.Rows[i][1].Value", sheet.Rows[i][userIdColumnId].Value)
		if sheet.Rows[i][wnumberColumnId].Value == wnumber || sheet.Rows[i][userIdColumnId].Value == wnumber { // search for our latest entry
			return i, true
		}
	}
	if sheet.Rows[prevI][wnumberColumnId].Value == wnumber {
		return prevI, true
	} else {
		return 0, false
	}
}

/*
func getLastEmptyRow(sheet *spreadsheet.Sheet) (int) {
	prevI := 0
	for i, val := range sheet.Rows {
		if val[0].Value == "" { // search for our latest entry
			return prevI
		}
		prevI = i
	}
	return prevI
}*/

func isLocaleColumn(sheet *spreadsheet.Sheet, key string, locale string) (bool) {
	for _, column := range sheet.Rows[0] {
		if column.Value == key+"_"+locale { // Cm format locali
			log.Println("Localed column: ", key, "; locale:", locale)
			return true
		}
	}
	log.Println("NOT localed column: ", key, "; locale:", locale)
	return false
}

func findColumn(sheet *spreadsheet.Sheet, key string) (int) {
	out := 0
	for i, column := range sheet.Rows[0][5:] {
		if column.Value == key {
			out = i
		}
	}
	return 5 + out
}

func findColumnNoPadding(sheet *spreadsheet.Sheet, key string) (int) {
	out := 0
	for i, column := range sheet.Rows[0] {
		if column.Value == key {
			out = i
		}
	}
	return out
}

func fillUnfilledCols(sheet *spreadsheet.Sheet, row int, lastColumnIdx int) {
	for i := 0; i <= lastColumnIdx-1; i++ {
		if sheet.Rows[row][i].Value == "" {
			sheet.Update(row, i, "0")
		}
	}
}

func fillColumn(sheet *spreadsheet.Sheet, columnIdx int, emptyRowIdx int) {
	//log.Println("Filling column(length = ",emptyRowIdx,") ",columnIdx)
	for i := 0; i <= emptyRowIdx; i++ {
		sheet.Update(i, columnIdx, "0")
	}
	//log.Println("Done!")
}

func fillWholeSheet(sheet *spreadsheet.Sheet, emptyRowIdx int, emptyColumnIdx int) {
	//numeric := regexp.MustCompile(`^([0-9]+)$`)
	for j := 0; j <= emptyColumnIdx-1; j++ { //j - column
		for i := 1; i <= emptyRowIdx-1; i++ { //i - row
			/*if !numeric.MatchString(sheet.Rows[i][j].Value) && sheet.Rows[i][j].Value != "not markable" && sheet.Rows[0][j].Value == "mark" {
				sheet.Update(i, j, "not markable")
			} else*//*
			if sheet.Rows[i][j].Value == "0" {
				if j > 4 {
					sheet.Update(i, j, "0")
				} else {
					sheet.Update(i, j, "0")
				}
			} else */if sheet.Rows[i][j].Value == "" {
				sheet.Update(i, j, "0")
			}
		}
	}
}

func checkErrorWithPush(err error, serviceId string, userTableTitle string) {
	if err != nil {
		log.Println("ERROR: " + err.Error())
		if serviceId != "" {
			log.Println("Pushing err with SID: " + serviceId)
			pushApi.pushErr(serviceId, err, userTableTitle)
		}
	}
}

func checkError(err error) {
	if err != nil {
		log.Println("ERROR: " + err.Error())
	}
}
