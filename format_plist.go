package main

import (
	"container/list"
	"encoding/xml"
	"image"
	"os"
	"strconv"
	"strings"
)

func resolveFramesFromPlist(plist string) *list.List {

	parsePoints := func(s string) []image.Point {
		ps := []image.Point{{}, {}, {}, {}}
		strs := strings.Split(s, ",")
		ps[0].X, _ = strconv.Atoi(strings.Replace(strs[0], "{", "", -1))
		ps[0].Y, _ = strconv.Atoi(strings.Replace(strs[1], "}", "", -1))
		if len(strs) == 4 {
			ps[1].X, _ = strconv.Atoi(strings.Replace(strs[2], "{", "", -1))
			ps[1].Y, _ = strconv.Atoi(strings.Replace(strs[3], "}", "", -1))
		}
		return ps
	}

	l := list.New()
	file, _ := os.Open(plist)
	defer file.Close()
	decoder := xml.NewDecoder(file)
	currentTagName := ""
	currentKey := ""
	var frame *Frame
	for token, _ := decoder.Token(); token != nil; token, _ = decoder.Token() {
		switch element := token.(type) {
		case xml.StartElement:
			currentTagName = element.Name.Local
			if currentTagName == "dict" {
				frame = &Frame{right: false}
				frame.key = currentKey
			}
			if currentKey == "rotated" || currentKey == "textureRotated" {
				if currentTagName == "true" {
					frame.rotated = true
				} else if currentTagName == "false" {
					frame.rotated = false
				}
			}
		case xml.EndElement:
			if element.Name.Local == "dict" && frame != nil && frame.sourceSize.X > 0 {
				if frame.spriteOffset.X != 0 || frame.spriteOffset.Y != 0 {
					frame.sourceOffset.X = frame.sourceSize.X/2 + frame.spriteOffset.X - frame.frameSize.X/2
					frame.sourceOffset.Y = frame.sourceSize.Y/2 - frame.spriteOffset.Y - frame.frameSize.Y/2
				}
				l.PushBack(frame)
			}
		case xml.CharData:
			s := string(element)
			s = strings.TrimSpace(s)
			if s == "" {
				break
			}
			if currentTagName == "key" {
				currentKey = s
				break
			}
			switch currentKey {
			case "frame":
				ps := parsePoints(s)
				frame.frameOffset = ps[0]
				frame.frameSize = ps[1]
			case "sourceColorRect":
				ps := parsePoints(s)
				frame.sourceOffset = ps[0]
			case "sourceSize":
				ps := parsePoints(s)
				frame.sourceSize = ps[0]
			case "spriteSize":
				ps := parsePoints(s)
				frame.frameSize = ps[0]
			case "spriteSourceSize":
				ps := parsePoints(s)
				frame.sourceSize = ps[0]
			case "textureRect":
				ps := parsePoints(s)
				frame.frameOffset = ps[0]
			case "spriteOffset":
				p := parsePoints(s)[0]
				// if frame.rotated {
				// frame.sourceOffset.X = frame.sourceSize.X/2 - p.X - frame.frameSize.X/2
				// frame.sourceOffset.Y = frame.sourceSize.Y/2 - p.Y - frame.frameSize.Y/2
				// } else {
				// frame.sourceOffset.X = frame.sourceSize.X/2 - p.X - frame.frameSize.X/2
				// frame.sourceOffset.Y = frame.sourceSize.Y/2 - p.Y - frame.frameOffset.Y/2
				// }
				frame.spriteOffset = p
			}
		}
	}
	return l
}
