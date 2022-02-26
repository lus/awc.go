package awc

import "fmt"

type endpoint string

const (
	endpointMETAR endpoint = "https://aviationweather.gov/adds/dataserver_current/httpparam?dataSource=metars&requestType=retrieve&format=xml"
)

func (end endpoint) addString(key, value string) endpoint {
	return endpoint(fmt.Sprintf("%s&%s=%s", end, key, value))
}

func (end endpoint) addBool(key string, value bool) endpoint {
	return endpoint(fmt.Sprintf("%s&%s=%t", end, key, value))
}

func (end endpoint) addInt(key string, value int64) endpoint {
	return endpoint(fmt.Sprintf("%s&%s=%d", end, key, value))
}

func (end endpoint) addFloat(key string, value float32) endpoint {
	return endpoint(fmt.Sprintf("%s&%s=%f", end, key, value))
}

func (end endpoint) String() string {
	return string(end)
}
