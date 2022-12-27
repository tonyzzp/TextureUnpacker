package main

import (
	"container/list"
	"fmt"
	"image"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

func resolveFramesFromAtlas(path string) *list.List {
	l := list.New()
	bytes, _ := ioutil.ReadFile(path)
	content := string(bytes)
	lines := strings.Split(content, "\n")
	lines = lines[1:]
	var frame *Frame
	for _, line := range lines {
		if strings.Index(line, ":") == -1 {
			if frame != nil {
				l.PushBack(frame)
			}
			frame = &Frame{right: true}
			frame.key = strings.TrimSpace(line) + ".png"
			continue
		}
		if frame == nil {
			continue
		}
		strs := strings.Split(strings.TrimSpace(line), ":")
		if len(strs) != 2 {
			fmt.Println("atlas格式错误", line)
			os.Exit(-1)
		}
		key := strings.TrimSpace(strs[0])
		val := strings.TrimSpace(strs[1])
		p := image.Pt(0, 0)
		if index := strings.Index(val, ","); index > -1 {
			p.X, _ = strconv.Atoi(strings.TrimSpace(val[:index]))
			p.Y, _ = strconv.Atoi(strings.TrimSpace(val[index+1:]))
		}
		switch key {
		case "rotate":
			frame.rotated = val == "true"
		case "xy":
			frame.frameOffset = p
		case "size":
			frame.frameSize = p
		case "orig":
			frame.sourceSize = p
		case "offset":
			frame.sourceOffset.X = p.X
			frame.sourceOffset.Y = frame.sourceSize.Y - p.Y - frame.frameSize.Y
		}
	}
	if frame != nil && frame.sourceSize.X > 0 {
		l.PushBack(frame)
	}
	return l
}
