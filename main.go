package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"golang.org/x/net/html"
)

/*
 * go run main.go http://lorjula.com http://schier.co https://google.com
 */
func main() {

	// Arguments
	seedUrls := os.Args[1:]

	foundUrls := make(map[string]bool)

	// channals
	chUrls := make(chan string)
	chaFinish := make(chan bool)

	for _, url := range seedUrls {
		go extrackHref(url, chUrls, chaFinish)
	}

	// Subscribe to both channels
	for s := 0; s < len(seedUrls); {
		select {
		case url := <-chUrls:
			foundUrls[url] = true
		case <-chaFinish:
			s++
		}
	}

	// We're done! Print the results...
	fmt.Println("\nFound", len(foundUrls), "unique urls:\n")

	for url, _ := range foundUrls {
		fmt.Println(" - " + url)
	}

	close(chUrls)
}

func extrackHref(url string, chUrls chan string, chaFinish chan bool) {

	// Request
	resp, err := http.Get(url)

	defer func() {
		// Notify that we're done after this function
		chaFinish <- true
	}()

	if err != nil {
		fmt.Println("ERROR: Failed to crawl \"" + url + "\"")
		return
	}

	// fmt.Println("Html:", toString(resp.Body))

	body := resp.Body
	defer body.Close()

	tag := html.NewTokenizer(body)

	for {

		next := tag.Next()

		switch {
		case next == html.ErrorToken:
			// End of the document, we're done
			return
		case next == html.StartTagToken:
			tk := tag.Token()
			if isAnchor := tk.Data == "a"; !isAnchor {
				continue
			}

			// Extract the href value, if there is one
			ok, url := href(tk)
			if !ok {
				continue
			}

			// Make sure the url begines in http**
			if found := strings.Index(url, "http"); found == 0 {
				chUrls <- url
			}

		}

	}

}

func href(token html.Token) (ok bool, ref string) {
	for _, a := range token.Attr {
		if a.Key == "href" {
			ref = a.Val
			ok = true
		}
	}
	return
}

func toString(body io.Reader) string {
	bytes, err := ioutil.ReadAll(body)

	if err != nil {
		fmt.Println("Error: ", err)
	}

	return string(bytes)
}
