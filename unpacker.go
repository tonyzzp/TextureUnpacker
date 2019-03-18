package main

import (
	"container/list"
	"encoding/xml"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type _Frame struct {
	key          string
	rotated      bool
	right        bool
	frameOffset  image.Point
	frameSize    image.Point
	sourceSize   image.Point
	sourceOffset image.Point
}

func isFile(path string) bool {
	st, _ := os.Stat(path)
	return st != nil && !st.IsDir()
}

func parsePoints(s string) []image.Point {
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

func resolveFramesFromPlist(plist string) *list.List {
	l := list.New()
	file, _ := os.Open(plist)
	defer file.Close()
	decoder := xml.NewDecoder(file)
	currentTagName := ""
	currentKey := ""
	var frame *_Frame
	for token, _ := decoder.Token(); token != nil; token, _ = decoder.Token() {
		switch element := token.(type) {
		case xml.StartElement:
			currentTagName = element.Name.Local
			if currentTagName == "dict" {
				frame = &_Frame{right: false}
				frame.key = currentKey
			}
			if currentKey == "rotated" {
				if currentTagName == "true" {
					frame.rotated = true
				} else if currentTagName == "false" {
					frame.rotated = false
				}
			}
		case xml.EndElement:
			if element.Name.Local == "dict" && frame != nil && frame.sourceSize.X > 0 {
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
			}
		}
	}
	return l
}

func resolveFramesFromAtlas(path string) *list.List {
	l := list.New()
	bytes, _ := ioutil.ReadFile(path)
	content := string(bytes)
	lines := strings.Split(content, "\n")
	lines = lines[1:]
	var frame *_Frame
	for _, line := range lines {
		if strings.Index(line, ":") == -1 {
			if frame != nil {
				l.PushBack(frame)
			}
			frame = &_Frame{right: true}
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

func rotateImage(img image.Image, right bool) *image.RGBA {
	w := img.Bounds().Dy()
	h := img.Bounds().Dx()
	rgba := image.NewRGBA(image.Rect(0, 0, w, h))
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			var c color.Color
			if right {
				c = img.At(y, w-x)
			} else {
				c = img.At(h-y, x)
			}
			rgba.Set(x, y, c)
		}
	}
	return rgba
}

func main() {
	_imgPath := flag.String("image", "", "png/jpg图片路径")
	flag.Parse()
	inPath, _ := filepath.Abs(*_imgPath)
	{
		st, _ := os.Stat(inPath)
		if st == nil || st.IsDir() {
			fmt.Println("拆分TexturePacker打包的拼图为单图。支持plist format=2")
			flag.Usage()
			return
		}
	}
	dir := filepath.Dir(inPath)
	baseName := filepath.Base(inPath)
	baseName = baseName[:strings.Index(baseName, ".")]
	outDir := filepath.Join(dir, baseName+"_out")

	var frames *list.List
	if path := filepath.Join(dir, baseName+".plist"); isFile(path) {
		frames = resolveFramesFromPlist(path)
	} else if path := filepath.Join(dir, baseName+".atlas"); isFile(path) {
		frames = resolveFramesFromAtlas(path)
	}

	file, _ := os.Open(inPath)
	source, _, _ := image.Decode(file)
	os.Mkdir(outDir, 0777)
	for e := frames.Front(); e != nil; e = e.Next() {
		frame := e.Value.(*_Frame)
		fmt.Println(frame)
		tw := frame.frameSize.X
		th := frame.frameSize.Y
		if frame.rotated {
			tw, th = th, tw
		}
		tmp := image.NewRGBA(image.Rect(0, 0, tw, th))
		draw.Draw(tmp, tmp.Rect, source, frame.frameOffset, draw.Src)
		if frame.rotated {
			tmp = rotateImage(tmp, frame.right)
		}
		dst := image.NewRGBA(image.Rect(0, 0, frame.sourceSize.X, frame.sourceSize.Y))
		rect := image.Rectangle{}
		rect.Min = frame.sourceOffset
		rect.Max = image.Pt(rect.Min.X+frame.frameSize.X, rect.Min.Y+frame.frameSize.Y)
		draw.Draw(dst, rect, tmp, image.Pt(0, 0), draw.Src)
		outFile := filepath.Join(outDir, frame.key)
		out, _ := os.Create(outFile)
		png.Encode(out, dst)
	}
	fmt.Println("成功 总数量", frames.Len())
}
