package cfinstanceinfoapi

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type AppInfo struct {
	Name  string `json:"name,omitempty"`
	Guid  string `json:"guid,omitempty"`
	Space string `json:"space,omitempty"`
	Org   string `json:"org,omitempty"`
}

func UpdateAppMap(apiUrl string, appmap map[string]AppInfo) {
	c := time.Tick(3 * time.Minute)
	for _ = range c {
		GenAppMap(apiUrl, appmap)
	}
}

func GenAppMap(apiUrl string, appmap map[string]AppInfo) {
	log.Println("updating app map")

	// get space info from cf-portal url
	pres, err := http.Get(apiUrl)
	if err != nil {
		log.Fatal(err)
	}

	// read portal url response body
	pbody, err := ioutil.ReadAll(pres.Body)
	pres.Body.Close()
	if err != nil {
		log.Fatal(err)
	}

	var pinfo []AppInfo
	err = json.Unmarshal(pbody, &pinfo)

	//fmt.Printf("%+v",pinfo)

	for index := range pinfo {
		appmap[pinfo[index].Guid] = pinfo[index]
	}

	//fmt.Printf("%v",appmap)
}
