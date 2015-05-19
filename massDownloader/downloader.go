package main

import (
	"flag"
	"fmt"
	"github.com/captncraig/mosaicChallenge/imgur"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"regexp"
	"sync"
)

var reddit = flag.String("sub", "", "Subreddit to fetch images for")
var output = flag.String("out", "", "Output directory for downloaded images.")
var count = flag.Int("n", 2000, "Number of images to fetch")
var designSeeds = flag.Bool("ds", true, "Download all images from designseeds.")

const workers = 20

type download struct{ url, local string }

var workChan = make(chan *download)

var waitGroup = sync.WaitGroup{}

func init() {
	for i := 0; i < workers; i++ {
		go func() {
			for {
				dl := <-workChan

				outFile := path.Clean(path.Join(*output, dl.local))
				if _, err := os.Stat(outFile); os.IsNotExist(err) {
					fmt.Println(dl.url, dl.local)
					file, err := os.Create(outFile)
					if err != nil {
						log.Fatal(err)
					}
					resp, err := http.Get(dl.url)
					if err != nil {
						log.Fatal(err)
					}
					_, err = io.Copy(file, resp.Body)
					if err != nil {
						log.Fatal(err)
					}
					resp.Body.Close()
					file.Close()

				}
				waitGroup.Done()
			}
		}()
	}
}

func main() {
	flag.Parse()
	*reddit = "cats"
	*output = "../collections/designseeds"

	files := []*download{}
	if *output == "" {
		fmt.Println("Output directory required")
		flag.PrintDefaults()
		return
	}

	if *designSeeds {
		var err error
		files, err = getDesignSeedsImages()
		if err != nil {
			log.Fatal(err)
		}
	} else {
		if *reddit == "" {
			fmt.Println("SubReddit required")
			flag.PrintDefaults()
			return
		}
		c := imgur.NewClient(nil)
		ids, err := c.GetTopSubredditImages(*reddit, *count)
		if err != nil {
			log.Fatal(err)
		}
		for _, id := range ids {
			files = append(files, &download{fmt.Sprintf("http://imgur.com/%ss.png", id), id + "s.png"})
		}
	}
	waitGroup.Add(len(files))
	for _, dl := range files {
		workChan <- dl
	}
	waitGroup.Wait()
}

func getDesignSeedsImages() ([]*download, error) {
	files := []*download{}
	rgx := regexp.MustCompile(`href="(.+\.png)"`)
	for _, url := range []string{"http://design-seeds.com/palettes/", "http://design-seeds.com/palettes/12/"} {
		resp, err := http.Get(url)
		if err != nil {
			return nil, err
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		matches := rgx.FindAllStringSubmatch(string(body), -1)
		for _, m := range matches {
			files = append(files, &download{url + m[1], m[1]})
		}
	}
	return files, nil
}
