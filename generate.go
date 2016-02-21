// +build ignore

//
// This file generates "gotags.go", which contains Go tags and sha hash releases.
//

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/format"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/govend/goversion/linkheader"
)

// Tag represents GitHub tag via the v3 API.
// More details at https://developer.github.com/v3/repos/#list-tags.
type Tag struct {
	Name   string
	Commit Commit
}

type Commit struct {
	SHA string
	URL string
}

func main() {

	tags := &[]Tag{}
	DownloadTags("https://api.github.com/repos/golang/go/tags", tags)

	// write preliminary file data such as comments, package name, structs, etc..
	var buf bytes.Buffer
	buf.WriteString("// this file is auto-generated by generate.go\n\n")
	buf.WriteString("package goversion\n")
	buf.WriteString(`
  type Tag struct {
  	Name   string
  	Commit Commit
  }

  type Commit struct {
  	SHA string
  	URL string
  }
	`)

	// write the dynamic list of standard packages
	fmt.Fprintf(&buf, "var gotags = %#v\n", tags)

	// transfer buffer bytes to final source
	src := buf.Bytes()

	// replace main.pkg type name with pkg
	src = bytes.Replace(src, []byte("main.Tag"), []byte("Tag"), -1)
	src = bytes.Replace(src, []byte("main.Commit"), []byte("Commit"), -1)

	// format all the source bytes
	src, err := format.Source(src)
	if err != nil {
		log.Fatal(err)
	}

	// write source bytes to the "stdpkgs.go" file
	if err := ioutil.WriteFile("gotags.go", src, 0644); err != nil {
		log.Fatal(err)
	}
}

// DownloadTags downloads all tag info and marshales it into a slice of
// GitHubTag.
func DownloadTags(url string, tags *[]Tag) error {

	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	t := &[]Tag{}
	if err := json.Unmarshal(body, t); err != nil {
		fmt.Println(string(body))
		panic(err)
	}
	*tags = append(*tags, *t...)

	links, err := linkheader.Parse(resp.Header.Get("Link"))
	if err != nil {
		panic(err)
	}

	for _, link := range links {
		if link.Rel == "next" {
			if err := DownloadTags(link.URI, tags); err != nil {
				panic(err)
			}
		}
	}

	return nil
}
