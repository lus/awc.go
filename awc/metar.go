package awc

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"time"
)

// METARQuery represents the query used to fetch METAR objects.
// Please keep in mind that a call either to HoursBeforeNow or Between is required.
// Please refer to https://aviationweather.gov/dataserver/example?datatype=metar for further information.
type METARQuery struct {
	station                                        *string
	startTime, endTime                             *int64
	hoursBeforeNow                                 *float32
	mostRecent                                     *bool
	mostRecentForEachStation                       *string
	rectMinLat, rectMinLon, rectMaxLat, rectMaxLon *float32
	radRadius, radLat, radLon                      *float32
	fields                                         []string
}

// Station specifies the station string to use for METAR querying
func (query *METARQuery) Station(value string) *METARQuery {
	query.station = &value
	return query
}

// Between specifies a timespan to fetch the METAR(s) in.
// If HoursBeforeNow was used before, that will be ignored.
func (query *METARQuery) Between(start, end time.Time) *METARQuery {
	startUnix := start.Unix()
	endUnix := end.Unix()

	query.startTime = &startUnix
	query.endTime = &endUnix

	query.hoursBeforeNow = nil

	return query
}

// HoursBeforeNow specifies the amount of hours before the current timestamp to fetch the METAR(s) from.
// If Between was used before, that will be ignored.
func (query *METARQuery) HoursBeforeNow(value float32) *METARQuery {
	value = float32(math.Abs(float64(value)))

	query.hoursBeforeNow = &value

	query.startTime = nil
	query.endTime = nil

	return query
}

// MostRecent specifies whether to only include the most recent METAR.
// If MostRecentForEachStation was used before, that will be ignored.
func (query *METARQuery) MostRecent(value bool) *METARQuery {
	query.mostRecent = &value

	query.mostRecentForEachStation = nil

	return query
}

// MostRecentForEachStation specifies the value for the 'mostRecentForEachStation' constraint.
// If MostRecent was used before, that will be ignored.
func (query *METARQuery) MostRecentForEachStation(value string) *METARQuery {
	query.mostRecentForEachStation = &value

	query.mostRecent = nil

	return query
}

// InRectangle specifies a rectangle consisting of min/max latitude and longitude to fetch the METAR(s) from.
// If RadialDistance was used before, that will be ignored.
func (query *METARQuery) InRectangle(minLat, minLon, maxLat, maxLon float32) *METARQuery {
	minLat = keepFloatInRange(minLat, -90, 90)
	minLon = keepFloatInRange(minLon, -180, 180)
	maxLat = keepFloatInRange(maxLat, -90, 90)
	maxLon = keepFloatInRange(maxLon, -180, 180)

	query.rectMinLat = &minLat
	query.rectMinLon = &minLon
	query.rectMaxLat = &maxLat
	query.rectMaxLon = &maxLon

	query.radRadius = nil
	query.radLat = nil
	query.radLon = nil

	return query
}

// RadialDistance specifies a radial distance consisting of latitude, longitude and radius to fetch the METAR(s) from.
// If InRectangle was used before, that will be ignored.
func (query *METARQuery) RadialDistance(radius, lat, lon float32) *METARQuery {
	radius = keepFloatInRange(radius, 0, 500)
	if radius == 0 {
		radius = 1
	}
	lat = keepFloatInRange(lat, -90, 90)
	lon = keepFloatInRange(lon, -180, 180)

	query.radRadius = &radius
	query.radLat = &lat
	query.radLon = &lon

	query.rectMinLat = nil
	query.rectMinLon = nil
	query.rectMaxLat = nil
	query.rectMaxLon = nil

	return query
}

// Fields specifies a list of fields to limit the response to
func (query *METARQuery) Fields(values ...string) *METARQuery {
	query.fields = values
	return query
}

func (query *METARQuery) buildEndpoint() endpoint {
	end := endpointMETAR
	if query.station != nil {
		end = end.addString("stationString", *query.station)
	}
	if query.startTime != nil {
		end = end.addInt("startTime", *query.startTime).addInt("endTime", *query.endTime)
	}
	if query.hoursBeforeNow != nil {
		end = end.addFloat("hoursBeforeNow", *query.hoursBeforeNow)
	}
	if query.mostRecent != nil {
		end = end.addBool("mostRecent", *query.mostRecent)
	}
	if query.mostRecentForEachStation != nil {
		end = end.addString("mostRecentForEachStation", *query.mostRecentForEachStation)
	}
	if query.rectMinLat != nil {
		end = end.
			addFloat("minLat", *query.rectMinLat).
			addFloat("minLon", *query.rectMinLon).
			addFloat("maxLat", *query.rectMaxLat).
			addFloat("maxLon", *query.rectMaxLon)
	}
	if query.radRadius != nil {
		end = end.addString("radialDistance", fmt.Sprintf("%f;%f,%f", *query.radRadius, *query.radLon, *query.radLat))
	}
	if len(query.fields) > 0 {
		end = end.addString("fields", strings.Join(query.fields, ","))
	}
	return end
}

// METARResponse represents the response that gets sent by the AWC Text Data Server
type METARResponse struct {
	XMLName  xml.Name `xml:"response"`
	Errors   []string `xml:"errors>error"`
	Warnings []string `xml:"warnings>warning"`
	METARs   []*METAR `xml:"data>METAR"`
}

// METAR represents a single METAR information object
type METAR struct {
	RawText                   string                   `xml:"raw_text"`
	StationID                 string                   `xml:"station_id"`
	ObservationTime           string                   `xml:"observation_time"`
	Latitude                  float32                  `xml:"latitude"`
	Longitude                 float32                  `xml:"longitude"`
	AirTempC                  float32                  `xml:"temp_c"`
	DewPointC                 float32                  `xml:"dewpoint_c"`
	WindDirDegrees            int                      `xml:"wind_dir_degrees"`
	WindSpeedKT               int                      `xml:"wind_speed_kt"`
	WindGustKT                int                      `xml:"wind_gust_kt"`
	VisibilityStatuteMI       float32                  `xml:"visibility_statute_mi"`
	AltimeterInHG             float32                  `xml:"altim_in_hg"`
	SeaLevelPressureMB        float32                  `xml:"sea_level_pressure_mb"`
	QualityControlFlags       METARQualityControlFlags `xml:"quality_control_flags"`
	WXString                  string                   `xml:"wx_string"`
	SkyConditions             []METARSkyCondition      `xml:"sky_condition"`
	FlightCategory            string                   `xml:"flight_category"`
	ThreeHRPressureTendencyMB float32                  `xml:"three_hr_pressure_tendency_mb"`
	MaxAirTemp6HC             float32                  `xml:"maxT_c"`
	MinAirTemp6HC             float32                  `xml:"minT_c"`
	MaxAirTemp24HC            float32                  `xml:"maxT24hr_c"`
	MinAirTemp24HC            float32                  `xml:"minT24hr_c"`
	PrecipitationIN           float32                  `xml:"precip_in"`
	Precipitation3HIN         float32                  `xml:"pcp3hr_in"`
	Precipitation6HIN         float32                  `xml:"pcp6hr_in"`
	Precipitation24HIN        float32                  `xml:"pcp24hr_in"`
	SnowDepthIN               float32                  `xml:"snow_in"`
	VerticalVisibilityFT      int                      `xml:"vert_vis_ft"`
	METARType                 string                   `xml:"metar_type"`
	ElevationM                float32                  `xml:"elevation_m"`
}

// METARQualityControlFlags contains the different METAR quality control flags
type METARQualityControlFlags struct {
	Corrected               bool `xml:"corrected"`
	Auto                    bool `xml:"auto"`
	AutoStation             bool `xml:"auto_station"`
	MaintenanceIndicator    bool `xml:"maintenance_indicator"`
	NoSignal                bool `xml:"no_signal"`
	LightningSensorOff      bool `xml:"lightning_sensor_off"`
	FreezingRainSensorOff   bool `xml:"freezing_rain_sensor_off"`
	PresentWeatherSensorOff bool `xml:"present_weather_sensor_off"`
}

// METARSkyCondition represents a single METAR sky condition entry
type METARSkyCondition struct {
	SkyCover       string `xml:"sky_cover,attr"`
	CloudBaseFTAGL int    `xml:"cloud_base_ft_agl,attr"`
}

// GetMETAR executes a METARQuery.
// Please keep in mind that this method only returns an error if the request itself failed or the server responded with
// a non-successful (code < 200 || code > 299) status code.
// The returned METARResponse contains separate fields that contain warnings and errors due to the AWC Text Data Server
// design.
func GetMETAR(query *METARQuery) (*METARResponse, error) {
	httpResponse, err := http.Get(query.buildEndpoint().String())
	if err != nil {
		return nil, err
	}
	if httpResponse.StatusCode < 200 || httpResponse.StatusCode > 299 {
		return nil, errors.New(fmt.Sprintf("unexpected status code: %d", httpResponse.StatusCode))
	}

	defer httpResponse.Body.Close()
	body, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return nil, err
	}

	response := new(METARResponse)
	if err := xml.Unmarshal(body, response); err != nil {
		return nil, err
	}

	return response, nil
}
