package main

import (
	"container/list"
	"encoding/json"
	"image"
	"io/ioutil"
)

func resolveFramesFromJson(path string) *list.List {
	getInt := func(data map[string]interface{}, key string) int {
		v := data[key]
		f := v.(float64)
		return int(f)
	}

	getFrameFromJsonObject := func(data map[string]interface{}, item *Frame) {
		frame := data["frame"].(map[string]interface{})
		item.frameOffset = image.Pt(getInt(frame, "x"), getInt(frame, "y"))
		item.frameSize = image.Pt(getInt(frame, "w"), getInt(frame, "h"))

		item.rotated = data["rotated"].(bool)

		spriteSourceSize := data["spriteSourceSize"].(map[string]interface{})
		item.sourceOffset = image.Pt(getInt(spriteSourceSize, "x"), getInt(spriteSourceSize, "y"))
		item.sourceSize = image.Pt(getInt(spriteSourceSize, "w"), getInt(spriteSourceSize, "h"))

		sourceSize := data["sourceSize"].(map[string]interface{})
		item.sourceSize = image.Pt(getInt(sourceSize, "w"), getInt(sourceSize, "h"))
	}

	l := list.New()
	bytes, _ := ioutil.ReadFile(path)
	m := make(map[string]interface{})
	json.Unmarshal(bytes, &m)
	if frames := m["frames"]; frames != nil {
		switch t := frames.(type) {
		case map[string]interface{}:
			for k, v := range t {
				item := Frame{}
				item.key = k
				getFrameFromJsonObject(v.(map[string]interface{}), &item)
				l.PushBack(&item)
			}
		case []interface{}:
			for _, v := range t {
				m := v.(map[string]interface{})
				item := Frame{}
				item.key = m["filename"].(string)
				getFrameFromJsonObject(m, &item)
				l.PushBack(&item)
			}
		}
	} else if textures := m["textures"]; textures != nil {
		frames := textures.([]interface{})[0].(map[string]interface{})["frames"].([]interface{})
		for _, v := range frames {
			m := v.(map[string]interface{})
			item := Frame{}
			item.key = m["filename"].(string)
			getFrameFromJsonObject(m, &item)
			l.PushBack(&item)
		}
	}
	return l
}
