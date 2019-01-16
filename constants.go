package main

const (
	ApiPattern       string = "https://test2.tamtam.chat/%s?access_token=<access_tocken>"
	ApiGeoPattern    string = "https://geocode-maps.yandex.ru/1.x/?apikey=<api_key>&format=json&geocode=%s"
	ApiRoutePattern  string = "https://taxi-routeinfo.taxi.yandex.net/taxi_info?clid=<cid>&apikey=<api_key>&rll=%s,%s~%s,%s"
	RouteLinkPattern string = "https://3.redirect.appmetrica.yandex.com/route?start-lat=%s&start-lon=%s&end-lat=%s&end-lon=%s&ref=bot.tam&appmetrica_tracking_id=25395763362139037"

	HelpMessage string = "== Yandex.Taxi help bot ==\n" +
		"usage:\n" +
		"--help  помоги мне!\n" +
		"--route {адрес откуда}:{адрес куда}  инфа о маршруте\n" +
		"--aliasroute {alias}  инфа о маршруте по алиасу\n" +
		"--alias {адрес откуда}:{адрес куда}:{alias}  сохранить алиас для маршрута\n" +
		"--aliaslist  список алиасов\n"
)
