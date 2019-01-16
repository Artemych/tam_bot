package main

import "sync"

type UpdatesResponse struct {
	Updates []Update `json:"updates"`
}

type GeoCodeResponse struct {
	Response struct {
		GeoObjectCollection struct {
			FeatureMembers []FeatureMember `json:"featureMember"`
		} `json:"GeoObjectCollection"`
	} `json:"response"`
}

type FeatureMember struct {
	GeoObject struct {
		Point struct {
			Pos string `json:"pos"`
		} `json:"Point"`
	} `json:"GeoObject"`
}

type Point struct {
	Lat string
	Lon string
}

type Update struct {
	UpdateType string `json:"update_type"`
	Message    struct {
		Sender struct {
			UserID int64  `json:"user_id"`
			Name   string `json:"name"`
		} `json:"sender"`
		Recipient struct {
			ChatID int64 `json:"chat_id"`
		} `json:"recipient"`
		Timestamp int64 `json:"timestamp"`
		Message   struct {
			Mid  string `json:"mid"`
			Seq  int64  `json:"seq"`
			Text string `json:"text"`
		} `json:"message"`
	} `json:"message"`
	Timestamp int64 `json:"timestamp"`
}

type MessageBody struct {
	Text string `json:"text"`
}

type LinkMessageBody struct {
	Text       string           `json:"text"`
	Attachment []LinkAttachment `json:"attachments"`
}

type LinkAttachment struct {
	Type    string         `json:"type"`
	Payload PayloadContent `json:"payload"`
}

type PayloadContent struct {
	Buttons [][]PayloadButton `json:"buttons"`
}

type PayloadButton struct {
	Type   string `json:"type"`
	Text   string `json:"text"`
	Intent string `json:"intent"`
	Url    string `json:"url"`
}

type RouteInfoResponse struct {
	Currency string  `json:"currency"`
	Distance float64 `json:"distance"`
	Options  []struct {
		ClassLevel  int     `json:"class_level"`
		ClassName   string  `json:"class_name"`
		ClassText   string  `json:"class_text"`
		MinPrice    float64 `json:"min_price"`
		Price       float64 `json:"price"`
		PriceText   string  `json:"price_text"`
		WaitingTime float64 `json:"waiting_time"`
	} `json:"options"`
	Time float64 `json:"time"`
}

type Alias struct {
	AddressFrom string
	AddressTo   string
	PointFrom   Point
	PointTo     Point
}

type Aliases struct {
	sync.Mutex
	storage map[string]Alias
}
