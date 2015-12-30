package transitspark

import (
	"fmt"
	"encoding/xml"
	"io/ioutil"
	"net/http"
)


type Client struct {
	key          string
	http         *http.Client
}



type Result struct {
    AgencyList AgencyList `xml:"AgencyList"`
}

type AgencyList struct {
    Agency Agency `xml:"Agency"`
}

type Agency struct {
    Name string `xml:"Name,attr"`
    RouteList RouteList `xml:"RouteList"`
}

type RouteList struct {
    Route []Route `xml:Route"`
}

type Route struct {
    Name string `xml:"Name,attr"`
    Code string `xml:"Code,attr"`
    RouteDirectionList []RouteDirectionList 
}

type RouteDirectionList struct {
    RouteDirection RouteDirection
}

type RouteDirection struct {
    Name string `xml:"Name,attr"`
    Code string `xml:"Code,attr"`
    StopList StopList
}


type StopList struct {
    Stop []Stop
}

type Stop struct {
    Name string `xml:"name,attr"`
    StopCode string `xml:"StopCode,attr"`
    DepartureTimeList  DepartureTimeList
}

type DepartureTimeList struct {
    DepartureTime []string `xml:"DepartureTime"`
}


func getRoutes(c *Client) *RouteList {

	resp, err := c.http.Get(fmt.Sprintf("http://services.my511.org/Transit2.0/GetRoutesForAgencies.aspx?token=%s&agencyNames=%s", c.key, "SF-MUNI"))
	if err != nil {
	    fmt.Println(err)
		return nil
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
	    fmt.Println(err)
		return nil
	}	
	
	v := Result{}
    xml.Unmarshal(data, &v)
    return &v.AgencyList.Agency.RouteList
}


func getStops(c *Client, route string, direction string) *RouteList {

	resp, err := c.http.Get(fmt.Sprintf("http://services.my511.org/Transit2.0/GetStopsForRoutes.aspx?token=%s&routeIDF=%s~%s~%s", c.key, "SF-MUNI", route, direction))
	if err != nil {
	    fmt.Println(err)
		return nil
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
	    fmt.Println(err)
		return nil
	}	
	
	v := Result{}
    xml.Unmarshal(data, &v)
    return &v.AgencyList.Agency.RouteList
}


func getNextDepartures(c *Client, stopCode string) *RouteList {

	resp, err := c.http.Get(fmt.Sprintf("http://services.my511.org/Transit2.0/GetNextDeparturesByStopCode.aspx?token=%s&agencyName=%s&stopCode=%s", c.key, "SF-MUNI", stopCode))
	if err != nil {
	    fmt.Println(err)
		return nil
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
	    fmt.Println(err)
		return nil
	}	
	
	v := Result{}
    xml.Unmarshal(data, &v)
    return &v.AgencyList.Agency.RouteList
}


