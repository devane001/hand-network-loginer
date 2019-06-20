package loginer

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/tidwall/gjson"
)

type Loginer struct {
	Username     string
	Password     string
	HeartBeatCyc int64
	SerialNo     int64
}

const (
	LoginURL     = "http://192.168.211.101/portal/pws?t=li"
	HeartBeatURL = "http://192.168.211.101/portal/page/doHeartBeat.jsp"
)

func (loginer *Loginer) Login() {
	client := &http.Client{}
	log.Printf(`login by user: %v`, loginer.Username)
	req, err := http.NewRequest("POST", LoginURL, strings.NewReader(fmt.Sprintf(`userName=%s&userPwd=%s`, loginer.Username, loginer.Password)))
	if err != nil {
		log.Panic(err)
	}
	req.Header.Set("Accept", "text/html")
	req.Header.Set("Accept-Language", "en-US")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		log.Panic(err)
	}
	defer resp.Body.Close()
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Panic(err)
	}
	respJSONBytes, err := base64.RawStdEncoding.DecodeString(string(respBytes))
	if err != nil {
		log.Panic(err)
	}
	respJSON, err := url.QueryUnescape(string(respJSONBytes))
	if err != nil {
		log.Panic(err)
	}
	loginer.HeartBeatCyc = gjson.Get(respJSON, "heartBeatCyc").Int()
	loginer.SerialNo = gjson.Get(respJSON, "serialNo").Int()
	if loginer.HeartBeatCyc == 0 || loginer.SerialNo == 0 {
		log.Panic(`login failed`, respJSON)
	}
	log.Printf(`login success serial number: %v`, loginer.SerialNo)
}

func (loginer *Loginer) HeartBeat() {
	client := &http.Client{}
	log.Printf(`heart beat by serial number: %v`, loginer.SerialNo)
	req, err := http.NewRequest("POST", HeartBeatURL, strings.NewReader(fmt.Sprintf(`serialNo=%v`, loginer.SerialNo)))
	if err != nil {
		log.Panic(err)
	}
	req.Header.Set("Accept", "text/html")
	req.Header.Set("Accept-Language", "en-US")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		log.Panic(err)
	}
	defer resp.Body.Close()
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Panic(err)
	}
	if strings.Index(string(respBytes), "parent.v_failedTimes=0;") == -1 {
		log.Panic("heart beat failed.", string(respBytes))
	}
	log.Printf(`heart beat success by serial number: %v`, loginer.SerialNo)
}
