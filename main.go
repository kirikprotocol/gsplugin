package main

import (
	"fmt"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"log"
	"time"
	"strconv"
	"os"
	"io"
	"strings"
	"net/url"
	"regexp"
)

type Config struct {
	ServerRoot          string
	Port                string
	UnmarkableXML       string
	MarkableXML         string
	ErrorXML            string
	LastEntryXML        string
	GetLastEntryPrefix  string
	LogPath             string
	PathToGoogleKeyJson string
	GoogleUsername      string
	GooglePassword      string
	KnownKeys           []string
}

type RedirectUrl struct {
	Url      string
	Caller   string
	Redirect bool
}

type User struct {
	Name  string
	Phone string
	Email string
}

var config = new(Config)
//var outputFile = new(os.File)
var responseXml = []byte{}
var markableResponseXml = []byte{}
var errorXml = []byte{}
var lastEntryResponseXml = []byte{}
var pushApi = PushApi{
	RequestTransport: &http.DefaultTransport}

// added for test commit

//var knownKeys = []string{"ref_sid", "event.id", "event.order", "subscriber", "abonent", "protocol", "user_id", "service", "event.text", "event.referer", "event", "lang", "serviceId", "wnumber"}

func init_system() (*Config, []byte, []byte, []byte, []byte, error) {
	cfg_bytes, err := ioutil.ReadFile(os.Args[1])
	json.Unmarshal(cfg_bytes, config)
	//log.Println("config: ",config)
	/*
	if !exists("out.csv") {
		ioutil.WriteFile("out.csv", []byte("page,button,user_id,wnumber,protocol\n"), 0644)
	}
	*/
	//f, err := os.OpenFile("out.csv", os.O_APPEND|os.O_WRONLY, 0600)
	resp_xml, err := ioutil.ReadFile(config.UnmarkableXML)
	mark_resp_xml, err := ioutil.ReadFile(config.MarkableXML)
	err_xml, err := ioutil.ReadFile(config.ErrorXML)
	last_entry_xml, err := ioutil.ReadFile(config.LastEntryXML)

	if err != nil {
		log.Fatal("Error reading from response files: ", err.Error())
	}

	logFile, err := os.OpenFile(config.LogPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer logFile.Close()
	mw := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(mw)
	log.Println("Logging to file and console!")

	initialize_sheet()
	return config, resp_xml, err_xml, mark_resp_xml, last_entry_xml, err
}

func calcMark(params map[string]string) (int) {
	out := 0
	regex := "^([0-9]+)$"
	for _, value := range params {
		matched, err := regexp.MatchString(regex, value)
		if matched && err == nil && !contains(config.KnownKeys, value) {
			log.Println("Matched: "+value)
			num, err := strconv.Atoi(value)
			if err == nil {
				out += num
			} else if value == "0" {
				out += 0
			}
		}else{
			log.Println("Error in calc mark: ",err)
		}
	}
	return out
}

func getLastHandler(w http.ResponseWriter, r *http.Request) { // parameters: list of parameters to return separated by comma
	//sample req: localhost:8800/getLast/?spreadsheetId=1GCXT5ii2NJxok6hpnjAXp3RQd6H_9TQs4pkKB6PDbZc&parameters=timestamp
	if len(r.URL.Query()) == 0 {
		fmt.Fprintf(w, string(errorXml), "Empty request!")
		return
	}
	log.Println("Got add request", r.URL.String())
	var lkServiceId = r.URL.Query().Get("serviceId")
	var userTableTitle = r.URL.Query().Get("userTableTitle")
	var tableTitle = r.URL.Query().Get("tableTitle")
	var wnumber = r.URL.Query().Get("wnumber")
	if wnumber == ""{
		wnumber = r.URL.Query().Get("subscriber")
	}
	if userTableTitle == "" {
		userTableTitle = "DispatchPhoneList"
	}
	var callback = r.URL.Query().Get("callback")
	var locale = r.URL.Query().Get("lang")
	var translationTableTitle = r.URL.Query().Get("translationTableTitle")
	if translationTableTitle == "" {
		translationTableTitle = "Translation" //Default name
	}
	var values = []string{}
	log.Println("Callback url: ", callback)
	updErr := updSheet(r.URL.Query().Get("spreadsheetId")) // Id of spreadsheet should be passed in "spreadsheetId" parameter
	if updErr != nil {
		fmt.Fprintf(w, string(errorXml), parseUpdErr(string(updErr.Error())))
		return
	}
	log.Println("Locale:", locale)
	response := config.GetLastEntryPrefix
	if locale != "" {
		log.Println("Getting localed result")
		values = getLastEntry(strings.Split(r.URL.Query().Get("parameters"), ","), wnumber, lkServiceId, userTableTitle, tableTitle)
		if values == nil {
			response += "No last entry"
			goto afterCheck
		}
		parameterNames := getParameterNamesInOrder(append(strings.Split(r.URL.Query().Get("parameters"), ","), "prefix"), locale, translationTableTitle)
		if parameterNames[len(parameterNames)-1] != "prefix" { //Setup prefix
			response = parameterNames[len(parameterNames)-1]
		}
		response += "<br/>"
		for i, val := range parameterNames {
			if i != len(parameterNames)-1 { // While value is not prefix
				response += val + ": " + values[i] + " <br/>"
			}
		}
	} else {
		log.Println("Locale not passed or locale is en, getting default result")
		values = getLastEntry(strings.Split(r.URL.Query().Get("parameters"), ","), r.URL.Query().Get("wnumber"), lkServiceId, userTableTitle,tableTitle) // if there is no locale passed
		if values == nil {
			response += "No last entry"
			goto afterCheck
		}
		for i, val := range strings.Split(r.URL.Query().Get("parameters"), ",") {
			response += val + ": " + values[i] + " <br/>"
		}
	}
	log.Println("ret result: ", values)
afterCheck:

	if callback != "" {
		//log.Printf(string(lastEntryResponseXml), response, callback)
		n, err := fmt.Fprintf(w, string(lastEntryResponseXml), response, url.QueryEscape(callback))
		if err != nil {
			log.Fatal(err, "; ", n)
		}
	} else {
		fmt.Fprintf(w, string(responseXml), response) //return just response
	}
}

func parseUpdErr(err string) (string) {
	out := "Google sheet plugin error: "
	if strings.Contains(err, "404") {
		out += "google sheet not found(404)"
	} else if strings.Contains(err, "403") {
		out += "access denied!(403) You should share your sheet to miniapps@miniappstesterbot.iam.gserviceaccount.com"
	} else {
		out += "unknown google sheets error: " + err
	}
	return out
}

func mainHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Got request:", r.URL.String(), "\nContent: ", r.Body)
	if len(r.URL.Query()) == 0 {
		fmt.Fprintf(w, string(errorXml), "Empty request!")
		return
	}
	var lkServiceId = r.URL.Query().Get("serviceId")
	var userTableTitle = r.URL.Query().Get("userTableTitle")
	if userTableTitle == "" {
		userTableTitle = "DispatchPhoneList"
	}
	var mark = 0
	evaluableInt, err := strconv.Atoi(r.URL.Query().Get("evaluable")) // Evaluabelness of task should be passed in "evaluable" parameter
	evaluable := true
	evaluable = evaluableInt == 1
	if err != nil || evaluableInt > 1 {
		evaluable = false
	}
	//for parameter := range r.URL.Query() {
	//if contains(config.PageNames, parameter) {
	params := genParameters(r.URL.Query())
	mark = calcMark(params)
	updErr := updSheet(r.URL.Query().Get("spreadsheetId")) // Id of spreadsheet should be passed in "spreadsheetId" parameter
	if updErr != nil {
		fmt.Fprintf(w, string(errorXml), parseUpdErr(string(updErr.Error())))
		return
	}
	go addEntry(time.Now().String(),
		r.URL.Query().Get("subscriber"),
		r.URL.Query().Get("protocol"),
		r.URL.Query().Get("wnumber"),
		//r.URL.Query().Get("event.id"),
		evaluable,
		mark,
		params,
		lkServiceId,
		userTableTitle)
	//go outputFile.Write([]byte(time.Now().String() + "," +
	//	parameter + "," +
	//	r.URL.Query().Get(parameter) + "," +
	//	r.URL.Query().Get("subscriber") + "," +
	//	r.URL.Query().Get("wnumber") + "," +
	//	r.URL.Query().Get("protocol") + "\n"))
	//}
	//}
	dispatch := r.URL.Query().Get("dispatch") //Make phone list dispatch
	log.Println("dispatch ", dispatch)
	if dispatch == "1" {
		log.Println("dispatch = 1")
		// miniapps globalussd-lk id
		log.Println("Making a dispatch with serviceId: " + lkServiceId)

		var locale = r.URL.Query().Get("lang")
		if locale == "" {
			locale = "en"
		}
		var translationTableTitle = r.URL.Query().Get("translationTableTitle")
		if translationTableTitle == "" {
			translationTableTitle = "Translation" //Default name
		}
		prefix := getParameterNamesInOrder([]string{"pushPrefix"}, locale, translationTableTitle)[0]
		log.Println("Got prefix: "+prefix)
		if prefix == "" { // if prefix is null
			prefix = "New entry:" // Default push prefix
		}
		if lkServiceId != "" {
			userTableTitle := r.URL.Query().Get("userTableTitle")
			if r.URL.Query().Get("userTableTitle") == "" {
				userTableTitle = "DispatchPhoneList" // default table title
			}
			sendEmail := r.URL.Query().Get("sendEmail") == "1"
			log.Println("Making push...")
			go pushApi.makePush(
				lkServiceId,
				locale,
				translationTableTitle,
				prefix,
				userTableTitle, //title of table with admins
				r.URL.Query().Get("subscriber"),
				r.URL.Query().Get("protocol"),
				evaluable,
				mark,
				sendEmail,
				params)
		}
	}
	callback := r.URL.Query().Get("callback") // request should have "callback" parameter
	if callback == "" {
		if !evaluable {
			fmt.Fprintf(w, string(responseXml), "Thank you for participating!!!")
		} else {
			fmt.Fprintf(w, string(markableResponseXml), strconv.Itoa(mark))
		}
	} else {
		http.Redirect(w, r, callback, 302)
	}
}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Redirect hander: ", r.URL.Query().Get("subscriber"))
	/*
	wg.Wait()
	wg.Add(1)
	firstRedirect := <-redirectUrls
	if !(firstRedirect.Caller == r.URL.Query().Get("subscriber")) { // search for our user
		redirectUrls <- firstRedirect //Put first redirect back
		for redirect := range redirectUrls {
			log.Println("Current redirect: ",redirect)
			if redirect.Caller == r.URL.Query().Get("subscriber"){
				log.Println("Found our redirect url: ",redirect)
				log.Println(len(redirectUrls))
				wg.Done()
				//removeFromChannel(redirect, &redirectUrls)
				if redirect.Redirect {
					http.Redirect(w, r, redirect.Url, 302)
				}else {
					fmt.Fprintf(w, string(responseXml), "Go to start")
				}//return default "on start" page
				return
			}
		}
	}else { //if caller is our subscriber
		wg.Done()
		if firstRedirect.Redirect {
	*/
	http.Redirect(w, r, r.URL.Query().Get("callback"), 302)
	/*/	}else {
			fmt.Fprintf(w, string(responseXml), "Go to start")
		}//return default "on start" page
	}*/
}

func main() {
	log.Println("Starting...")
	if len(os.Args) < 2 {
		log.Fatal("You should pass me a config name like: ", os.Args[0], " <json config name>")
	}
	cfg, respXml, errXml, markRespXml, lastEntrResponseXml, err := init_system()
	config = cfg
	//outputFile = f
	errorXml = errXml
	responseXml = respXml
	markableResponseXml = markRespXml
	lastEntryResponseXml = lastEntrResponseXml
	//log.Println(string(response_xml))
	log.Println("Config: ", config)
	if err != nil {
		//outputFile.Close()
		panic(err)
	}
	log.Println("Done! Listening...")
	http.HandleFunc(config.ServerRoot, mainHandler)
	http.HandleFunc(config.ServerRoot+"getLast/", getLastHandler)
	http.HandleFunc(config.ServerRoot+"redirect/", redirectHandler)
	http.ListenAndServe(":"+config.Port, nil)
}
