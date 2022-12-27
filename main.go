package main

import (
	"container/list"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	"image/png"
	"os"
	"path/filepath"
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

	spriteOffset image.Point
}

func isFile(path string) bool {
	st, _ := os.Stat(path)
	return st != nil && !st.IsDir()
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
	} else if path := filepath.Join(dir, baseName+".txt"); isFile(path) {
		frames = resolveFramesFromAtlas(path)
	} else if path := filepath.Join(dir, baseName+".json"); isFile(path) {
		frames = resolveFramesFromJson(path)
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
