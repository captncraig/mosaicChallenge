package main

import (
	"fmt"
	"github.com/captncraig/mosaicChallenge/mosaics"
	"image"
	"image/gif"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"log"
	"os"
	"path"
	"time"
)

func main() {
	lib := mosaics.NewLibrary(mosaics.AveragingEvaluator())
	dirname := "collections/designSeeds"
	files, err := ioutil.ReadDir(dirname)
	if err != nil {
		log.Fatal(err)
	}
	for i, file := range files {
		fmt.Println(i, file.Name())
		img, _ := parseFile(dirname, file.Name())
		lib.AddImage(img)
	}
	f, err := os.Open("g.gif")
	if err != nil {
		log.Fatal(err)
	}

	g, err := gif.DecodeAll(f)
	if err != nil {
		log.Fatal(err)
	}
	start := time.Now()
	f.Close()
	//mos := mosaics.BuildMosaicFromLibrary(g.Image[0], lib, nil)
	out, err := mosaics.BuildGifzaic(g, lib, nil)
	output, _ := os.Create("moz.gif")
	fmt.Println(time.Now().Sub(start))
	start = time.Now()
	//fmt.Println(png.Encode(output, mos))
	//fmt.Println(jpeg.Encode(output, mos, &jpeg.Options{10}))
	fmt.Println(gif.EncodeAll(output, out))
	fmt.Println(time.Now().Sub(start))
	output.Close()
}

func parseFile(dir, file string) (image.Image, error) {
	f, err := os.Open(path.Join(dir, file))
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		log.Fatal(err)
	}

	return img, nil
}
