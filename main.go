package main

import (
    "html/template"
    "log"
    "net/http"
    "github.com/GeoNet/cfg"
    "fmt"
    "io/ioutil"
    "strings"
)

var templates map[string]*template.Template
var config = cfg.Load()

func init() {
    if templates == nil {
        templates = make(map[string]*template.Template)
    }
    templates["index"] = template.Must(template.ParseFiles("tmpl/base.tmpl", "tmpl/index.tmpl"))
    templates["edit"] = template.Must(template.ParseFiles("tmpl/base.tmpl", "tmpl/edit.tmpl"))

    //handle css files
    http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("css"))))
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
                Title: getFileTitle(name),
            })
        }
    }

    pageData := PageData{Title : "GeoNet News",
        AllMarkdown: allMarkdown}

    renderTemplate(w, "index", pageData)
}

func editPage(w http.ResponseWriter, r *http.Request) {
    v := r.URL.Query()
    mdName := v.Get("mdname")
    mdData := MarkdownData{}
    pageData := PageData{Title:"GeoNet News"}
    if mdName != "" {
        mdData.FileName = mdName
        mdData.Title = getFileTitle(mdName)
        b, err := readMarkDown(md_src_dir + mdName + ".md");
        if (err == nil) {
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

    log.Println("title ", title);

    content := r.Form["postContent"][0]

    log.Println("content ", content);
    //save md content
    saveMdContent(title, content)

    //TODO save html content

    //go to index page
    indexPage(w, r);
}


func renderTemplate(w http.ResponseWriter, name string, data interface{}) {
    tmpl, ok := templates[name]
    if (!ok) {
        http.Error(w, fmt.Sprintf("The template %s does not exist.", name), http.StatusInternalServerError)
    }

    err := tmpl.ExecuteTemplate(w, "base", &data)

    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}


type MarkdownData struct {
    Title       string
    FileName    string
    MdContent string
    HtmlContent   string
}

type PageData struct {
    Title       string
    AllMarkdown []MarkdownData
    MarkDown    MarkdownData
}

