package main

// testing script for Footlocker.com

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"
)

var (
	siteKey        = "288922D4BE1987530B4E5D4A17952C"
	previousCookie = ""
	step           = 0
)

// parts of payload edited
// TimeSinceLastEvent
// TimezoneOffset
// EvalLength
// TimeSeconds
// CookieLength
// mouse events

func CookieGen() string{
	payload, _ := os.ReadFile("payload.json")
	ddjsData := &fingerprint{}
	json.Unmarshal(payload, &ddjsData)

	timeNow := time.Now()

	ddjsData.TimeSinceLastEvent = rand.Float64() + float64(rand.Intn(450))
	_, tzoffset := timeNow.Zone()
	ddjsData.TimezoneOffset = tzoffset


	uaLower := strings.ToLower(ddjsData.Ua)
	if !strings.Contains(uaLower, "safari") && !strings.Contains(uaLower, "firefox") && !strings.Contains(uaLower, "other") {
		ddjsData.EvalLength = 37
	} else if !strings.Contains(uaLower, "internet Explorer") && !strings.Contains(uaLower, "other") {
		ddjsData.EvalLength = 39
	} else if !strings.Contains(uaLower, "chrome") && !strings.Contains(uaLower, "opera") && !strings.Contains(uaLower, "other") {
		ddjsData.EvalLength = 33
	}

	// curr time in ms
	ddjsData.TimeSeconds = (timeNow.UnixMilli() / 1000) - 1000


	data := &url.Values{}
	if step == 1 {
		// after 1st dd request
		ddjsData.CookieLength = 1
		e := &events{PageTimestamp: timeNow.UnixMilli(), Width: ddjsData.ArsW, Height: ddjsData.ArsH}
		e.MouseMoveEvents()
		ddjsData.EsSigmdn = e.Sigmdn
		ddjsData.EsMumdn = e.Mumdn
		ddjsData.EsDistmdn = e.Dismdn
		ddjsData.EsAngsmdn = e.Angsmdn
		ddjsData.EsAngemdn = e.Angemdn
		// below mouse vals only on footlocker
		ddjsData.MpCx = int(e.LastPosition.X)
		ddjsData.MpCy = int(e.LastPosition.Y)
		ddjsData.MpTr = true
		ddjsData.MpMx = &e.LastPosition.MovementX
		ddjsData.MpMy = &e.LastPosition.MovementY
		ddjsData.MpSx = e.LastPosition.ScreenX
		ddjsData.MpSy = e.LastPosition.ScreenY
		ddjsData.MmMd = e.WeirdMDValue
		f, _ := json.Marshal(ddjsData)
		data.Set("jsData", string(f))
		f, _ = json.Marshal(e)
		fmt.Println(string(f))
		data.Set("eventCounters", string(f))
		data.Set("jsType", "le")
		data.Set("cid", previousCookie)
		data.Set("ddk", siteKey)
	} else {
		// 1st dd request
		ddjsData.MpMx = nil
		ddjsData.MpMy = nil
		f, _ := json.Marshal(ddjsData)
		data.Set("jsData", string(f))
		data.Set("eventCounters", "[]")
		data.Set("jsType", "ch")
		data.Set("cid", "null")
		data.Set("ddk", siteKey)
		step++
	}
	data.Set("Referer", "https://www.footlocker.com/")
	data.Set("request", "/")
	data.Set("responsePage", "origin")
	data.Set("ddv", "4.6.12")

	jar, err := cookiejar.New(nil)
	if err != nil {
		panic(err)
	}
	client := &http.Client{
		Jar: jar,
	}

	req, err := http.NewRequest("POST", "https://api-js.datadome.co/js/", strings.NewReader(data.Encode()))
	req.Header.Add("accept", "*/*")
	req.Header.Add("accept-encoding", "gzip, deflate, br")
	req.Header.Add("accept-language", "en-US,en;q=0.9")
	req.Header.Add("content-type", "application/x-www-form-urlencoded")
	req.Header.Add("origin", "https://www.footlocker.com")
	req.Header.Add("referer", "https://www.footlocker.com/")
	req.Header.Add("sec-ch-ua", `"Not_A Brand";v="99", "Google Chrome";v="109", "Chromium";v="109\"`)
	req.Header.Add("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36")

	resp, err := client.Do(req)

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		panic(err)
	}
	re, _ := regexp.Compile("datadome=([^;]+)")
	ddCookie := re.FindStringSubmatch(string(body))[1]

	return ddCookie;
}

func (e *events) MouseMoveEvents() {
	// genning random browser mouse events
	mouseMovements := []*mouseMove{}
	// 2800-3200
	eventsNum := rand.Intn(400) + 2800
	e.MouseMove = eventsNum
	sigmas := []float64{}
	mumdn := []float64{}
	dists := []float64{}
	startAngles := []float64{}
	endAngles := []float64{}
	for idx := 0; idx < eventsNum; idx++ {
		mouseMovement := &mouseMove{}
		mouseMovement.X = float64(rand.Intn(e.Width))
		mouseMovement.Y = float64(rand.Intn(e.Height))
		mouseMovement.MovementX = rand.Intn(2) - 1 
		mouseMovement.MovementY = rand.Intn(2) - 1 
		mouseMovement.ScreenX = rand.Intn(e.Width)
		mouseMovement.ScreenY = rand.Intn(e.Height)
		// probably a bit high to be realistic but idc
		mouseMovement.Timestamp = rand.Float64() + float64(rand.Intn(450))
		mouseMovements = append(mouseMovements, mouseMovement)
		if idx%1000 == 0 || idx == eventsNum-1 {
			if idx == eventsNum-1 {
				e.LastPosition = mouseMovement
			}
			if e.WeirdMD == nil || e.WeirdMD.Timestamp == 0.0 {
				e.WeirdMD = mouseMovement
			} else {
				mouseMovementCalcs := math.Sqrt((mouseMovement.X-e.WeirdMD.X)*(mouseMovement.X-e.WeirdMD.X) + (mouseMovement.Y-e.WeirdMD.Y)*(mouseMovement.Y-e.WeirdMD.Y))
				if int(mouseMovementCalcs) > e.WeirdMDValue {
					e.WeirdMDValue = int(mouseMovementCalcs)
				}
			}
			sig, mu, dist, startAngle, endAngle := mouseEventFunc(mouseMovements[idx/1000:])
			sigmas = append(sigmas, sig)
			mumdn = append(mumdn, mu)
			dists = append(dists, dist)
			startAngles = append(startAngles, startAngle)
			endAngles = append(endAngles, endAngle)
		}
	}
	e.Sigmdn = returnMidPointArr(sigmas)
	e.Mumdn = returnMidPointArr(mumdn)
	e.Dismdn = returnMidPointArr(dists)
	e.Angsmdn = returnMidPointArr(startAngles)
	e.Angemdn = returnMidPointArr(endAngles)
	e.Click = rand.Intn(30)
	e.Scroll = rand.Intn(10)
	e.Keydown = rand.Intn(20)
	e.Keyup = e.Keydown
}

func mouseEventFunc(movements []*mouseMove) (float64, float64, float64, float64, float64) {
	// mimick dd script
	allTimeStampsLog := 0.0
	timesMouseMoved := len(movements)
	allTimeStampsLogSqed := 0.0
	for _, movement := range movements {
		movementTime := math.Log(movement.Timestamp)
		allTimeStampsLog += movementTime
		allTimeStampsLogSqed += movementTime * movementTime
	}
	mouseEvent1 := movements[0]
	mouseEvent2 := movements[timesMouseMoved - 1]
	mouseX := float64(mouseEvent1.X)
	mouseY := float64(mouseEvent1.Y)
	previousMouseX := float64(mouseEvent2.X)
	previousMouseY := float64(mouseEvent2.Y)
	differenceX := previousMouseX - mouseX
	differenceY := previousMouseY - mouseY
	lastIndex := 3
	if timesMouseMoved < 4 {
		lastIndex = timesMouseMoved - 1
	}
	lastEvent := movements[lastIndex]
	firstMouseEvent := movements[timesMouseMoved - lastIndex -1]

	return math.Sqrt( (float64(timesMouseMoved) * allTimeStampsLogSqed - allTimeStampsLog * allTimeStampsLog) / float64(timesMouseMoved) * (float64(timesMouseMoved)-1)) / 10000000.0,
	allTimeStampsLog / float64(timesMouseMoved),
	math.Sqrt(differenceX * differenceX + differenceY * differenceY),
	findMouseCoordsAngle(mouseEvent1.X, mouseEvent1.Y, lastEvent.X, lastEvent.Y),
	findMouseCoordsAngle(mouseEvent2.X, mouseEvent2.Y, firstMouseEvent.X, firstMouseEvent.Y)
}

func findMouseCoordsAngle(x1, y1, x2, y2 float64) float64 {
	// mimick dd script
	distBetweenXs := x2 - x1
	distBetweenYs := y2 - y1
	coordsAngle := math.Acos(distBetweenXs / math.Sqrt(distBetweenXs * distBetweenXs + (distBetweenYs * distBetweenYs)))
	if distBetweenYs < 0 {
		return -coordsAngle
	}
	return coordsAngle
}

func returnMidPointArr(arr []float64) float64 {
	// mimick dd script
	sort.Float64s(arr)
	halfway := float64((len(arr) - 1) * 50 / 100)
	halfwayIndex := int(math.Floor(halfway))
	if arr[halfwayIndex + 1] != 0 {
		almostZ := halfway - float64(halfwayIndex)
		return arr[halfwayIndex] + float64(almostZ)*(arr[halfwayIndex+1]-arr[halfwayIndex])
	}
	return arr[halfwayIndex]
}