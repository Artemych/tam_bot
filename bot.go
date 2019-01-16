package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

var httpClient = &http.Client{}
var aliases = &Aliases{
	storage: make(map[string]Alias),
}

func main() {
	fmt.Println("== Yandex.Taxi bot started ==")
	updatesResponse := &UpdatesResponse{}
	for {
		err := getJson(requestUrl("updates", nil), updatesResponse)

		if err != nil {
			log.Println(err)
			continue
		}

		if len(updatesResponse.Updates) == 0 {
			fmt.Println("no updates")
			continue
		}

		for _, update := range updatesResponse.Updates {
			if update.UpdateType != "message_created" {
				fmt.Println(update)
				continue
			}

			message := update.Message.Message.Text
			chatId := update.Message.Recipient.ChatID
			userId := update.Message.Sender.UserID

			fmt.Printf("New message recieved. chat:%d, sender:%d, messase:%s", chatId, userId, message)
			fmt.Println()

			handleMessage(chatId, userId, message)
		}
	}
}

func handleMessage(chatId int64, userId int64, message string) {
	if len(message) == 0 || !strings.HasPrefix(message, "--") {
		sendHelp(chatId)
		return
	}

	if strings.HasPrefix(message, "--help") {
		sendHelp(chatId)
		return
	}

	if strings.HasPrefix(message, "--route") {
		go handleRouteMessage(chatId, message)
		return
	}

	if strings.HasPrefix(message, "--aliasroute") {
		go handleAliasRouteMessage(chatId, userId, message)
		return
	}

	if strings.HasPrefix(message, "--aliaslist") {
		go handleListAliasMessage(chatId, userId)
		return
	}

	if strings.HasPrefix(message, "--alias") {
		go handleAliasMessage(chatId, userId, message)
		return
	}

	sendHelp(chatId)
}

func handleListAliasMessage(chatId int64, userId int64) {
	var buffer bytes.Buffer
	aliases.Lock()
	for k, v := range aliases.storage {
		if strings.HasPrefix(k, makeAliasPrefix(chatId, userId)) {
			buffer.WriteString(strings.Split(k, ":")[2] + " " + v.AddressFrom + " - " + v.AddressTo)
			buffer.WriteString("\n")
		}
	}
	aliases.Unlock()
	if len(buffer.Bytes()) == 0 {
		sendMessage(chatId, "Пусто!")
	} else {
		sendMessage(chatId, buffer.String())
	}
}

func handleRouteMessage(chatId int64, message string) {
	message = strings.Replace(message, "--route", "", 1)
	message = strings.TrimSpace(message)
	addresses := strings.Split(message, ":")

	if len(addresses) != 2 {
		sendMessage(chatId, fmt.Sprintf("Неверный аргумент команды --route %s", message))
		return
	}

	pointFrom, err := getGeoCodePoint(replaceSpaces(addresses[0]))

	if err != nil {
		sendMessage(chatId, fmt.Sprintf("Адрес %s не найден", addresses[0]))
		return
	}

	pointTo, err := getGeoCodePoint(replaceSpaces(addresses[1]))

	if err != nil {
		sendMessage(chatId, fmt.Sprintf("Адрес %s не найден", addresses[1]))
		return
	}

	requestRouteInfo(pointFrom, pointTo, chatId, addresses)
}

func requestRouteInfo(pointFrom *Point, pointTo *Point, chatId int64, addresses []string) {
	routeInfoResponse := &RouteInfoResponse{}
	err := getJson(getRouteUrl(pointFrom.Lon, pointFrom.Lat, pointTo.Lon, pointTo.Lat), routeInfoResponse)
	if err != nil || len(routeInfoResponse.Options) == 0 {
		sendMessage(chatId, fmt.Sprintf("can't get route info for %s", addresses))
		return
	}
	className := routeInfoResponse.Options[0].ClassText
	cost := routeInfoResponse.Options[0].PriceText
	waitingTime := convertWaitingTime(routeInfoResponse.Options[0].WaitingTime)
	routeTime := convertWaitingTime(routeInfoResponse.Time)
	text := fmt.Sprintf("Класс: %s\nСтоимость: %s\nВремя ожидания: %s\nВремя поездки: %s",
		className, cost, waitingTime, routeTime)
	url := makeRouteUrl(pointFrom, pointTo)
	sendMessageWithLinkButton(chatId, text, url)
}

func handleAliasRouteMessage(chatId int64, userId int64, message string) {
	message = strings.Replace(message, "--aliasroute", "", 1)
	message = strings.TrimSpace(message)

	if len(message) == 0 {
		sendMessage(chatId, fmt.Sprintf("Неверный аргумент для --aliasroute:%s", message))
		return
	}

	aliases.Lock()
	val, ok := aliases.storage[makeAliasKey(chatId, userId, message)]
	aliases.Unlock()
	if ok {
		addresses := []string{
			val.AddressFrom,
			val.AddressTo,
		}
		requestRouteInfo(&val.PointFrom, &val.PointTo, chatId, addresses)
	} else {
		sendMessage(chatId, fmt.Sprintf("Такого нет:%s", message))
	}
}

func handleAliasMessage(chatId int64, userId int64, message string) {
	message = strings.Replace(message, "--alias", "", 1)
	message = strings.TrimSpace(message)
	addresses := strings.Split(message, ":")

	if len(addresses) != 3 {
		sendMessage(chatId, fmt.Sprintf("Неверный аргумент для --alias:%s", message))
		return
	}

	if len(addresses[2]) == 0 {
		sendMessage(chatId, fmt.Sprintf("Неверный аргумент для --alias:%s", message))
		return
	}

	pointFrom, err := getGeoCodePoint(replaceSpaces(addresses[0]))

	if err != nil {
		sendMessage(chatId, fmt.Sprintf("Адрес %s не найден", addresses[0]))
		return
	}

	pointTo, err := getGeoCodePoint(replaceSpaces(addresses[1]))

	if err != nil {
		sendMessage(chatId, fmt.Sprintf("Адрес %s не найден", addresses[1]))
		return
	}

	aliasKey := makeAliasKey(chatId, userId, addresses[2])

	aliases.Lock()
	aliases.storage[aliasKey] = Alias{
		PointFrom:   *pointFrom,
		PointTo:     *pointTo,
		AddressFrom: addresses[0],
		AddressTo:   addresses[1],
	}
	aliases.Unlock()

	sendMessage(chatId, fmt.Sprintf("Алиас %s сохранен", addresses[2]))
}

func makeAliasKey(chatId int64, userId int64, alias string) string {
	return makeAliasPrefix(chatId, userId) + ":" + alias
}

func makeAliasPrefix(chatId int64, userId int64) string {
	return strconv.FormatInt(chatId, 10) + ":" + strconv.FormatInt(userId, 10)
}

func makeRouteUrl(point *Point, point2 *Point) string {
	return fmt.Sprintf(RouteLinkPattern, point.Lat, point.Lon, point2.Lat, point2.Lon)
}

func convertWaitingTime(time float64) string {
	hours, hourMinutes := divmod(int64(time), 3600)
	minutes, _ := divmod(int64(time), 60)
	if hours != 0 {
		return fmt.Sprintf("~%dчас %dмин.", hours, (hourMinutes/60)+1)
	}
	return fmt.Sprintf("~%dмин.", minutes+1)
}

func divmod(numerator, denominator int64) (quotient, remainder int64) {
	quotient = numerator / denominator
	remainder = numerator % denominator
	return
}

func getRouteUrl(lonFrom string, latFrom string, lonTo string, latTo string) string {
	return fmt.Sprintf(ApiRoutePattern, lonFrom, latFrom, lonTo, latTo)
}

func getGeoCodePoint(address string) (point *Point, err error) {
	geoCodeResponse := &GeoCodeResponse{}
	err = getJson(makeGeoCodeUrl(address), geoCodeResponse)
	if err != nil {
		return nil, err
	}

	if len(geoCodeResponse.Response.GeoObjectCollection.FeatureMembers) == 0 {
		return nil, fmt.Errorf("no geocode found for %s", address)
	}

	pos := strings.Split(geoCodeResponse.Response.GeoObjectCollection.FeatureMembers[0].GeoObject.Point.Pos, " ")

	return &Point{Lat: pos[1], Lon: pos[0]}, nil
}

func makeGeoCodeUrl(address string) string {
	return fmt.Sprintf(ApiGeoPattern, address)
}

func replaceSpaces(input string) string {
	return strings.Replace(input, " ", "+", -1)
}

func sendHelp(chatId int64) {
	go sendMessage(chatId, HelpMessage)
}

func sendMessage(chatId int64, text string) {
	body := MessageBody{Text: text}
	sendMessageInternal(body, chatId)
}

func sendMessageWithLinkButton(chatId int64, text string, link string) {

	payload := PayloadContent{}
	payload.Buttons = [][]PayloadButton{}

	row := []PayloadButton{
		{
			Type:   "link",
			Text:   "Поехали!",
			Intent: "default",
			Url:    link,
		},
	}

	payload.Buttons = append(payload.Buttons, row)

	attachment := []LinkAttachment{
		{
			Type:    "inline_keyboard",
			Payload: payload,
		},
	}

	body := LinkMessageBody{
		Text:       text,
		Attachment: attachment,
	}
	sendMessageInternal(body, chatId)
}

func sendMessageInternal(body interface{}, chatId int64) {
	buffer := new(bytes.Buffer)
	err := json.NewEncoder(buffer).Encode(body)
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Println(string(buffer.Bytes()))
	sendUrl := requestUrl("messages", map[string]string{
		"chat_id": strconv.FormatInt(chatId, 10),
	})
	res, err := httpClient.Post(sendUrl, "application/json", buffer)
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Println("message sent")
	readSimpleResponse(err, res)
}

func readSimpleResponse(err error, res *http.Response) {
	plainBuff := make([]byte, 4096)
	n, err := res.Body.Read(plainBuff)
	if n == 0 && err != nil {
		log.Println(err)
		return
	}

	defer res.Body.Close()

	fmt.Printf("%s", plainBuff[:n])
	fmt.Println()
}

func requestUrl(method string, params map[string]string) string {
	url := fmt.Sprintf(ApiPattern, method)
	if params == nil || len(params) == 0 {
		return url
	}
	var buffer bytes.Buffer
	buffer.WriteString(url)
	for k, v := range params {
		buffer.WriteString("&" + k + "=" + v)
	}
	return buffer.String()
}

func getJson(url string, target interface{}) error {
	r, err := httpClient.Get(url)
	if err != nil {
		return err
	}

	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(target)
}
