package main

import (
	"os"
	"io/ioutil"
	"bytes"
	"net/url"

)

func exists(path string) (bool) {
	_, err := os.Stat(path)
	if err == nil { return true }
	if os.IsNotExist(err) { return false }
	return true
}

func file_contains(fileName string, what string) (bool, error) {
	data, err := ioutil.ReadFile(fileName)
	return bytes.Contains(data, []byte(what)), err
}

func contains(arr []string, elem string) (bool){
	for _, element := range arr{
		if element == elem{
			return true
		}
	}
	return false
}

func map2arr(mapping map[string]string)([]string, []string){
	out1 := []string{}
	out2 := []string{}
	for key, val := range mapping{
		out1 = append(out1, key)
		out2 = append(out2, val)
	}
	return out1,out2
}

func max(a []int)(int){
	maximum := 0
	for _,val:=range a{
		if val > maximum{
			maximum = val
		}
	}
	return maximum
}

func genParameters(query url.Values) (map[string]string){
	out := map[string]string{}
	for parmaeter := range query{
		if !contains(config.KnownKeys, parmaeter){
			out[parmaeter] = query.Get(parmaeter)
		}
	}
	return out
}