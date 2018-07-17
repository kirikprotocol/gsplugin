package main

import (
	"log"
	"net/url"
	"net/http"
	"strconv"
	"regexp"
	"strings"
)

type PushApi struct {
	RequestTransport *http.RoundTripper
}

func (p *PushApi) makePush(lkServiceId string, locale string, translationTableTitle string, prefix string, userTableTitle string, subscriber string, protocol string, evaluable bool, mark int, sendEmail bool, params map[string]string) {
	var admins = getPhoneList(userTableTitle)
	log.Println("Phone list in make push: ", admins)
	for _, admin := range admins {
		if admin.Name != "" && admin.Phone != "" {
			var response, mailResponse = p.makeResponse(admin, locale, translationTableTitle, prefix, subscriber, protocol, evaluable, mark, params)
			log.Println("Response is: " + response)
			go p.push(lkServiceId, admin, response, protocol)
			if sendEmail && admin.Email != "" {
				go sendMail(admin.Email, mailResponse, locale, translationTableTitle)
			}
		}
	}
}

func (p *PushApi) makeResponse(admin User, locale string, translationTableTitle string, prefix string, subscriber string, protocol string, evaluable bool, mark int, params map[string]string) (string, string) {
	r1 := regexp.MustCompile(`^page([0-9]+)`)
	var response = "<?xml version=\"1.0\" encoding=\"utf-8\"?><page version=\"2.0\"><div>"
	var mailResponse = ""
	// TODO: make response
	log.Println("Making response...")
	response = response + prefix + "<br/>"
	mailResponse = mailResponse + prefix + "\n"
	if evaluable {
		response = response + "Mark: " + strconv.Itoa(mark) + "<br/>"
		mailResponse = mailResponse + "Mark: " + strconv.Itoa(mark) + "\n"
	}
	mailPostfix := "Regards, Spreadsheet Plugin."
	if len(params) != 0 {
		//response = response + "Other parameters:<br/>"
		keys, vals := map2arr(params)
		keys = append(keys, "mailPostfix")
		translatedKeys :=  getParameterNamesInOrder(keys, locale, translationTableTitle)
		log.Println("Translated keys: ",translatedKeys)
		for i, key := range translatedKeys {
			if !r1.MatchString(key) && keys[i] != "mailPostfix" && vals[i] != "back" && vals[i] != "main" && !strings.Contains(strings.ToLower(vals[i]), "notpushable"){
				response = response + key + ": " + vals[i] + "<br/>"
				mailResponse = mailResponse + key + ": " + vals[i] + "\n"
			}else if keys[i] == "mailPostfix"{
				log.Println("Mail postfix: "+key)
				mailPostfix = key
			}
		}
	}else {
		paramNames := getParameterNamesInOrder([]string{"mailPostfix"}, locale, translationTableTitle)
		if len(paramNames) > 0 {
			mailPostfix = paramNames[0]
			log.Println("Mail postfix: " + mailPostfix)
		}
	}
	response = response + "</div><navigation><link pageId=\"\">/start</link></navigation></page>"
	mailResponse += "\n"+mailPostfix//Regards, Spreadsheet Plugin."
	return response, mailResponse
}

func (p *PushApi) pushErr(lkServiceId string, err error, userTableTitle string) {
	var admins = getPhoneList(userTableTitle)
	log.Println("Phone list in make push: ", admins)
	for _, admin := range admins {
		if admin.Phone != "" {
			p.push(lkServiceId,admin,parseUpdErr(string(err.Error())), "telegram")
		}
	}
}

func (p *PushApi) push(lkServiceId string, admin User, response string, protocol string) {
	// : make push
	//transport := &http.Transport{Proxy: ... }
	transport := p.RequestTransport
	client := &http.Client{
		Transport: *transport,
	}
	payload := url.Values{}
	payload.Add("service", lkServiceId)
	payload.Add("subscriber", admin.Phone)
	payload.Add("protocol", "telegram") // Using only telegram protocol(for now)
	payload.Add("scenario", "xmlpush")
	payload.Add("document", response)
	log.Println("DBG: sendMsg addr: ", "http://prod.globalussd.mobi/push?; body:"+payload.Encode())
	req, err := http.NewRequest("GET", "http://prod.globalussd.mobi/push?"+payload.Encode(), nil)
	resp, err := client.Do(req)
	log.Println("req: ", resp)
	if err != nil {
		log.Println("error when making request to miniapps", err, req)
	}
}
