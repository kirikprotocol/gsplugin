package main

import (
	"net/smtp"
	"log"
	"strings"
	"fmt"
)

func sendMail(to string, body string, lang string, translationTableTitle string){
	body = strings.Replace(body, "\n", "\r\n", -1)
	mailSubject := "New entry"
	translatedKeys := getParameterNamesInOrder([]string{"mailSubject"}, lang, translationTableTitle)
	if len(translatedKeys) > 0{
		mailSubject = translatedKeys[0]
	}
	header := "From: Miniapps spreadsheet plugin <spreadsheetplugin@gmail.com>\r\n"+
	"Content-Type: text/plain; charset=utf-8\r\n"+
	"Content-Transfer-Encoding: quoted-printable\r\n"+
	"Subject: "+mailSubject+"\r\n"
	log.Println("Sending email to "+to+" with body: "+strings.Replace(body, "\r\n", "(CR)", -1))
	auth := smtp.PlainAuth(
		"Miniapps Spreadsheet Plugin",
		config.GoogleUsername,
		config.GooglePassword,
		"smtp.gmail.com",
	)
	err := smtp.SendMail(
		"smtp.gmail.com:25",
		auth,
		config.GoogleUsername,
		[]string{to},
		[]byte(header+"\r\n"+encodeMsg(body)),
	)
	checkError(err)
}

func encodeMsg(msg string)(string){
	output:= strings.Replace(fmt.Sprintf("% x", msg), " ", "=", -1)
	log.Println("Encoded text: ="+output)
	return "="+output
}
