package main

import (
	"io/ioutil"
	"os"
	"strings"
)

const (
	md_src_dir  = "_source/"
	md_html_dir = "_html/"
	img_dir     = "_images/"
)

func init() {
	if _, err := os.Stat(md_src_dir); os.IsNotExist(err) {
		os.Mkdir(md_src_dir, 0764)
	}
	if _, err := os.Stat(md_html_dir); os.IsNotExist(err) {
		os.Mkdir(md_html_dir, 0764)
	}
	if _, err := os.Stat(img_dir); os.IsNotExist(err) {
		os.Mkdir(img_dir, 0764)
	}
}

func getFileNameForTitle(title string) string {
	fleName := strings.Replace(title, "/", "-", -1)
	fleName = strings.Replace(fleName, " ", "_", -1)
	//check existence
	return fleName
}

func getFileTitle(name string) string {
	title := strings.Replace(name, "_", " ", -1)
	//check existence

	return title
}

func saveMdContent(title string, content string) {
	fleName := getFileNameForTitle(title)

	ioutil.WriteFile(md_src_dir+fleName+".md", []byte(content), 0644)

	//check existence

	//return fleName
}

func readMarkDown(filePath string) ([]byte, error) {
	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		return b, err
	}
	return b, nil

}
