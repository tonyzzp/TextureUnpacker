package main

import (
	"container/list"
	"encoding/xml"
	"flag"
	"fmt"
	"image"
	"image/draw"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type _Frame struct {
	key          string
	rotated      bool
	frameOffset  image.Point
	frameSize    image.Point
	sourceSize   image.Point
	sourceOffset image.Point
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
				frame = &_Frame{}
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

func rotateImage(img image.Image) *image.RGBA {
	w := img.Bounds().Dy()
	h := img.Bounds().Dx()
	rgba := image.NewRGBA(image.Rect(0, 0, w, h))
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			c := img.At(h-y, x)
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
	plistPath := filepath.Join(dir, baseName+".plist")
	outDir := filepath.Join(dir, baseName+"_out")

	file, _ := os.Open(inPath)
	source, _, _ := image.Decode(file)
	frames := resolveFramesFromPlist(plistPath)
	os.Mkdir(outDir, 0777)
	for e := frames.Front(); e != nil; e = e.Next() {
		frame := e.Value.(*_Frame)
		tw := frame.frameSize.X
		th := frame.frameSize.Y
		if frame.rotated {
			tw, th = th, tw
		}
		tmp := image.NewRGBA(image.Rect(0, 0, tw, th))
		draw.Draw(tmp, tmp.Rect, source, frame.frameOffset, draw.Src)
		if frame.rotated {
			tmp = rotateImage(tmp)
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
