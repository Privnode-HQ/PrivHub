package common

var ssoClientNameByID = map[string]string{
	"ticket-v1":                "Privnode 支持",
	"h8kdFu9IQTHdMV7wxdcZpqFv": "佬友API",
}

func GetSSOClientName(clientID string) (string, bool) {
	name, ok := ssoClientNameByID[clientID]
	return name, ok
}

func IsSSOClientIDAllowed(clientID string) bool {
	_, ok := ssoClientNameByID[clientID]
	return ok
}
