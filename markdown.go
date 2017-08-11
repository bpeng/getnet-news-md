package main

import (
	"io/ioutil"
	"os"
	"strings"
)

const (
	md_src_dir  = "_source/"
	md_html_dir = "pages/"
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
	//get rid of offending chars
	fleName := strings.Replace(title, "/", "-", -1)
	fleName = strings.Replace(title, ":", "-", -1)
	//replace space
	fleName = strings.Replace(fleName, " ", "_", -1)
	//check existence
	return fleName
}

/**
 * get the title to display on the html page
 * and the link to be disabled
 */
func getTitleForPage(pageData *PageData, title string) {
	if strings.Contains(title, "photos") { //Group photos
		pageData.Title = "惠村爬友群-群相册"
		pageData.LinkDisable = "3"
	} else if strings.Contains(title, "About") { //About group
		pageData.Title = "惠村爬友群-简介"
		pageData.LinkDisable = "2"
	} else if strings.Contains(title, "Aware") { //Aware
		pageData.Title = "惠村爬友群-注意事项"
		pageData.LinkDisable = "4"
	} else if strings.Contains(title, "walk") { //normal
		pageData.Title = "走路通知-" + title
		pageData.LinkDisable = "0"
	}
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
