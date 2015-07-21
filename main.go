package main

import (
	"bufio"
	"fmt"
	"github.com/GeoNet/cfg"
	"github.com/russross/blackfriday"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

var templates map[string]*template.Template
var config = cfg.Load()
var htmlFlags int
var htmlExt int

func init() {
	if templates == nil {
		templates = make(map[string]*template.Template)
	}
	templates["index"] = template.Must(template.ParseFiles("tmpl/base.tmpl", "tmpl/index.tmpl"))
	templates["edit"] = template.Must(template.ParseFiles("tmpl/base.tmpl", "tmpl/edit.tmpl"))
	templates["preview"] = template.Must(template.ParseFiles("tmpl/preview.tmpl"))

	//handle css files
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("css"))))

	// set up html options
	htmlExt = 0
	htmlExt |= blackfriday.EXTENSION_NO_INTRA_EMPHASIS
	htmlExt |= blackfriday.EXTENSION_TABLES
	htmlExt |= blackfriday.EXTENSION_FENCED_CODE
	htmlExt |= blackfriday.EXTENSION_AUTOLINK
	htmlExt |= blackfriday.EXTENSION_STRIKETHROUGH
	htmlExt |= blackfriday.EXTENSION_SPACE_HEADERS

	htmlFlags = blackfriday.HTML_USE_XHTML |
		blackfriday.HTML_USE_SMARTYPANTS |
		blackfriday.HTML_SMARTYPANTS_FRACTIONS |
		blackfriday.HTML_SMARTYPANTS_LATEX_DASHES |
		blackfriday.HTML_FOOTNOTE_RETURN_LINKS |
		blackfriday.HTML_SMARTYPANTS_ANGLED_QUOTES

}

func main() {
	http.HandleFunc("/", indexPage)

	http.HandleFunc("/edit", editPage)

	http.HandleFunc("/save", savePage)

	log.Fatal(http.ListenAndServe(config.WebServer.Port, nil))
}

func indexPage(w http.ResponseWriter, r *http.Request) {
	//get md files
	files, err := ioutil.ReadDir(md_src_dir)
	if err != nil {
		log.Fatal(err)
	}
	allMarkdown := make([]MarkdownData, 0)
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".md") {
			name := strings.Replace(f.Name(), ".md", "", 1)
			allMarkdown = append(allMarkdown,
				MarkdownData{
					FileName: name,
					Title:    getFileTitle(name),
				})
		}
	}

	pageData := PageData{Title: "GeoNet News",
		AllMarkdown: allMarkdown}

	renderTemplate(w, "index", pageData)
}

func editPage(w http.ResponseWriter, r *http.Request) {
	v := r.URL.Query()
	mdName := v.Get("mdname")
	mdData := MarkdownData{}
	pageData := PageData{Title: "GeoNet News"}
	if mdName != "" {
		mdData.FileName = mdName
		mdData.Title = getFileTitle(mdName)
		b, err := readMarkDown(md_src_dir + mdName + ".md")
		if err == nil {
			mdData.MdContent = string(b)
		}
		pageData.MarkDown = mdData
	}
	renderTemplate(w, "edit", pageData)
}

func savePage(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotAcceptable)
	}
	title := r.Form["postTitle"][0]

	log.Println("title ", title)

	content := r.Form["postContent"][0]
	//log.Println("content ", content);
	preview := r.Form["preview"][0]

	if preview == "1" {
		log.Println("preview ", preview)
		previewPage(w, r, title, content)

	} else {
		//save md content
		saveMdContent(title, content)
		indexPage(w, r)
	}
}

func previewPage(w http.ResponseWriter, r *http.Request, title string, content string) {
	mdData := MarkdownData{}
	pageData := PageData{Title: "GeoNet News"}
	mdData.Title = title
	mdData.MdContent = content

	htmlRenderer := blackfriday.HtmlRenderer(htmlFlags, "", "")
	mdData.HtmlContent = template.HTML(blackfriday.Markdown([]byte(content), htmlRenderer, htmlExt))
	pageData.MarkDown = mdData

	//save html preview
	saveHtmlContent(title, pageData)
	//render
	renderTemplate(w, "edit", pageData)
}

//render html page
func renderTemplate(w http.ResponseWriter, name string, data interface{}) {
	tmpl, ok := templates[name]
	if !ok {
		http.Error(http.ResponseWriter(w), fmt.Sprintf("The template %s does not exist.", name), http.StatusInternalServerError)
	}
	err := tmpl.ExecuteTemplate(w, "base", &data)
	if err != nil {
		http.Error(http.ResponseWriter(w), err.Error(), http.StatusInternalServerError)
	}
}

//save html content to local directory
func saveHtmlContent(title string, data interface{}) {
	fleName := getFileNameForTitle(title)
	f, err := os.Create(md_html_dir + fleName + ".html")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	tmpl, ok := templates["preview"]
	if !ok {
		panic("The template preview does not exist.")
	}
	err = tmpl.ExecuteTemplate(w, "base", &data)
	if err != nil {
		panic(err)
	}
	w.Flush()
}

type MarkdownData struct {
	Title       string
	FileName    string
	MdContent   string
	HtmlContent template.HTML
}

type PageData struct {
	Title       string
	AllMarkdown []MarkdownData
	MarkDown    MarkdownData
}
