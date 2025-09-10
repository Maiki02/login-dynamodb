package models

type SendMessageWuzapi struct {
	Phone   string            `json:"Phone"`
	Body    string            `json:"Body"`
	Id      string            `json:"Id"`
	Context ContextInfoWuzapi `json:"contextInfo"`
}

type SetStateWuzapi struct {
	Phone string `json:"Phone"`
	State string `json:"State"`
}

type ContextInfoWuzapi struct {
	StanzaId    string `json:"StanzaId"`
	Participant string `json:"Participant"`
}

type MessageReceiveWuzapi struct {
	Event EventWuzapi `json:"event"`
	Type  string      `json:"type"`
}

type EventWuzapi struct {
	Info       InfoWuzapi       `json:"Info"`
	Message    MessageWuzapi    `json:"Message"`
	RawMessage RawMessageWuzapi `json:"RawMessage"`
}

type InfoWuzapi struct {
	Chat               string            `json:"Chat"`
	Sender             string            `json:"Sender"`
	IsFromMe           bool              `json:"IsFromMe"`
	IsGroup            bool              `json:"IsGroup"`
	BroadcastListOwner string            `json:"BroadcastListOwner"`
	ID                 string            `json:"ID"`
	ServerID           int               `json:"ServerID"`
	Type               string            `json:"Type"`
	PushName           string            `json:"PushName"`
	Timestamp          string            `json:"Timestamp"`
	Category           string            `json:"Category"`
	Multicast          bool              `json:"Multicast"`
	MediaType          string            `json:"MediaType"`
	Edit               string            `json:"Edit"`
	MsgBotInfo         MsgBotInfoWuzapi  `json:"MsgBotInfo"`
	MsgMetaInfo        MsgMetaInfoWuzapi `json:"MsgMetaInfo"`
	VerifiedName       interface{}       `json:"VerifiedName"`
	DeviceSentMeta     interface{}       `json:"DeviceSentMeta"`
}

type MsgBotInfoWuzapi struct {
	EditType              string `json:"EditType"`
	EditTargetID          string `json:"EditTargetID"`
	EditSenderTimestampMS string `json:"EditSenderTimestampMS"`
}

type MsgMetaInfoWuzapi struct {
	TargetID     string `json:"TargetID"`
	TargetSender string `json:"TargetSender"`
}

type MessageWuzapi struct {
	Conversation        string                   `json:"conversation"`
	ExtendedTextMessage string                   `json:"extendedTextMessage"`
	MessageContextInfo  MessageContextInfoWuzapi `json:"messageContextInfo"`
}

type MessageContextInfoWuzapi struct {
	DeviceListMetadata        interface{} `json:"deviceListMetadata"`
	DeviceListMetadataVersion int         `json:"deviceListMetadataVersion"`
}

type RawMessageWuzapi struct {
	Conversation        string                   `json:"conversation"`
	ExtendedTextMessage string                   `json:"extendedTextMessage"`
	MessageContextInfo  MessageContextInfoWuzapi `json:"messageContextInfo"`
}

/*
"{
	"event":
		{"Info":
			{"Chat":"5493416887794@s.whatsapp.net",
			"Sender":"5493416887794@s.whatsapp.net",
			"IsFromMe":false,
			"IsGroup":false,
			"BroadcastListOwner":"",
			"ID":"3ED575C8F36A51557895C2BD46688028",
			"ServerID":0,
			"Type":"text",
			"PushName":"Miqueas Gentile",
			"Timestamp":"2024-10-13T19:31:11-03:00",
			"Category":"",
			"Multicast":false,
			"MediaType":"",
			"Edit":"",
			"MsgBotInfo":{
				"EditType":"",
				"EditTargetID":"",
				"EditSenderTimestampMS":"0001-01-01T00:00:00Z"
			},
			"MsgMetaInfo":{
				"TargetID":"",
				"TargetSender":""
			},
			"VerifiedName":null,
			"DeviceSentMeta":null
		},
		"Message":{
			"conversation":"aaa",
			"messageContextInfo":{
				"deviceListMetadata":{
					"senderKeyHash":"enS0VcVKBOcNSw==",
					"senderTimestamp":1728333112,
					"recipientKeyHash":"Cpfh4CttNIuxDA==",
					"recipientTimestamp":1728857956
				},
				"deviceListMetadataVersion":2
			}
		},
		"IsEphemeral":false,
		"IsViewOnce":false,
		"IsViewOnceV2":false,
		"IsViewOnceV2Extension":false,
		"IsDocumentWithCaption":false,
		"IsLottieSticker":false,
		"IsEdit":false,
		"SourceWebMsg":null,
		"UnavailableRequestID":"",
		"RetryCount":0,
		"NewsletterMeta":null,
		"RawMessage":{
			"conversation":"aaa",
			"messageContextInfo":{
				"deviceListMetadata":{
					"senderKeyHash":"enS0VcVKBOcNSw==",
					"senderTimestamp":1728333112,
					"recipientKeyHash":"Cpfh4CttNIuxDA==",
					"recipientTimestamp":1728857956
				},
				"deviceListMetadataVersion":2
			}
		}
	},
	"type":"Message"
}"*/

/*
{Event:{Info:{Chat:5493413544755@s.whatsapp.net Sender:5493413544755@s.whatsapp.net IsFromMe:false IsGroup:false BroadcastListOwner: ID:534AE0E38B44B56BF5A6F78A349A344C ServerID:0 Type:text PushName:Alana Gentile Timestamp:2024-10-13T20:51:40-03:00 Category: Multicast:false MediaType: Edit: MsgBotInfo:{EditType: EditTargetID: EditSenderTimestampMS:0001-01-01T00:00:00Z} MsgMetaInfo:{TargetID: TargetSender:} VerifiedName:<nil> DeviceSentMeta:<nil>} Message:{Conversation:Hola otra vez MessageContextInfo:{DeviceListMetadata:map[recipientKeyHash:Cpfh4CttNIuxDA== recipientTimestamp:1.728857956e+09] DeviceListMetadataVersion:2}} RawMessage:{Conversation:Hola otra vez MessageContextInfo:{DeviceListMetadata:map[recipientKeyHash:Cpfh4CttNIuxDA== recipientTimestamp:1.728857956e+09] DeviceListMetadataVersion:2}}} Type:Message}
*/
