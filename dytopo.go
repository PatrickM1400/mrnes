package mrnes

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/iti/evt/evtm"
	"github.com/iti/evt/vrtime"
	"gopkg.in/yaml.v3"
)

type DytopoEventCfg struct {
	EventID     string
	EventType   string  //Add or Remove link
	Time        float64 //Absolute time of event
	Interfaces1 IntrfcDesc
	Interfaces2 IntrfcDesc
}

type DytopoCfg struct {
	Name   string
	Events []DytopoEventCfg
}

func ReadDytopoCfg(dytopoFileName string, useYAML bool, dict []byte) (*DytopoCfg, error) {
	var err error

	// read from the file only if the byte slice is empty
	// validate input file name
	if len(dict) == 0 {
		fileInfo, err := os.Stat(dytopoFileName)
		if os.IsNotExist(err) || fileInfo.IsDir() {
			msg := fmt.Sprintf("dynamic topology %s does not exist or cannot be read", dytopoFileName)
			fmt.Println(msg)

			return nil, errors.New(msg)
		}
		dict, err = os.ReadFile(dytopoFileName)
		if err != nil {
			return nil, err
		}
	}

	// dict has slice of bytes to process
	example := DytopoCfg{}

	// input path extension identifies whether we deserialized encoded json or encoded yaml
	if useYAML {
		err = yaml.Unmarshal(dict, &example)
	} else {
		err = json.Unmarshal(dict, &example)
	}

	if err != nil {
		return nil, err
	}

	return &example, nil
}

func handleIntrfcEvent(evtMgr *evtm.EventManager, context any, data any) any {
	// fmt.Println(data)
	var dytopoEvent DytopoEventCfg = data.(DytopoEventCfg)
	fmt.Println("Handle Interface Event Called at ", evtMgr.Time)
	// if dytopoEvent == nil {
	// 	fmt.Println("Dytopo event is nil")
	// }
	// fmt.Println(dytopoEvent.EventType)
	if dytopoEvent.EventType == "Add" {

		addInterface(&dytopoEvent.Interfaces1)
		addInterface(&dytopoEvent.Interfaces2)
		linkIntrfcStruct(&dytopoEvent.Interfaces1)
		linkIntrfcStruct(&dytopoEvent.Interfaces2)
		// fmt.Println("Linked Interface Structures")
	} else {
		// fmt.Println(dytopoEvent.Interface.Name)
		removeInterface(dytopoEvent.Interfaces1.Name)
		removeInterface(dytopoEvent.Interfaces2.Name)
	}
	// fmt.Println("")
	return nil
}

func LoadDytopo(dytopoFile string, evtMgr *evtm.EventManager, traceMgr *TraceManager) error {
	empty := make([]byte, 0)
	ext := path.Ext(dytopoFile)
	useYAML := (ext == ".yaml") || (ext == ".yml")

	dc, err := ReadDytopoCfg(dytopoFile, useYAML, empty)
	fmt.Println("Starting LoadDytopo")
	if err != nil {
		return err
	}

	for idx := 0; idx < len(dc.Events); idx++ {
		// fmt.Println(dc.Events[idx])
		offset := vrtime.SecondsToTime((dc.Events[idx].Time - evtMgr.CurrentSeconds()) / 1000)
		// fmt.Println(offset)
		evtMgr.Schedule(nil, dc.Events[idx], handleIntrfcEvent, offset)

	}
	fmt.Println("Finished loading dynamic events")

	return nil
}
