package main

import (
	"bufio"
	"fmt"
	"github.com/GeoNet/cfg"
	"github.com/russross/blackfriday"
	"html/template"
	"io"
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
	templates["index1"] = template.Must(template.ParseFiles("tmpl/base.tmpl", "tmpl/index1.tmpl"))
	templates["edit"] = template.Must(template.ParseFiles("tmpl/base.tmpl", "tmpl/edit.tmpl"))
	templates["preview"] = template.Must(template.ParseFiles("tmpl/preview.tmpl"))

	//handle css files
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("css"))))
	http.Handle("/_images/", http.StripPrefix("/_images/", http.FileServer(http.Dir("_images"))))

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

	http.HandleFunc("/preview", indexPagePreview)

	http.HandleFunc("/edit", editPage)

	http.HandleFunc("/save", savePage)

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func indexPagePreview(w http.ResponseWriter, r *http.Request) {
	pageData, err := getIndexPageData()
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotAcceptable)
	}

	//save the page to html
	saveHtmlContent("index", "index1", pageData)
	//copy css, images
	err = CopyDir("css/", "_html/css/")
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Directory copied")
	}
	err = CopyDir("_images/", "_html/_images/")
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Directory copied")
	}

	renderTemplate(w, "index1", pageData)
}

func indexPage(w http.ResponseWriter, r *http.Request) {
	pageData, err := getIndexPageData()
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotAcceptable)
	}

	renderTemplate(w, "index", pageData)
}

func getIndexPageData() (*PageData, error) {
	//get md files
	files, err := ioutil.ReadDir(md_src_dir)
	if err != nil {
		return nil, err
	}
	allMarkdown := make([]MarkdownData, 0)
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".md") {
			name := strings.Replace(f.Name(), ".md", "", 1)
			//log.Println("files filename: %s, name%s", f.Name(), name)
			allMarkdown = append(allMarkdown,
				MarkdownData{
					FileName: name,
					Title:    getFileTitle(name),
				})
		}
	}

	return &PageData{Title: "GeoNet News",
		AllMarkdown: allMarkdown}, nil
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
	err := r.ParseMultipartForm(100000)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotAcceptable)
	}

	title := r.Form["postTitle"][0]

	log.Println("title ", title)

	content := r.Form["postContent"][0]
	//log.Println("content ", content);
	preview := r.Form["preview"][0]
	//log.Println("preview ", preview)
	if preview == "1" { //preview
		//log.Println("preview ", preview)
		previewPage(w, r, title, content)
	} else if preview == "2" { //load images
		//get a ref to the parsed multipart form
		m := r.MultipartForm
		//get the *fileheaders
		files := m.File["imagefiles"]
		var imgContents string
		for i, _ := range files {
			//for each fileheader, get a handle to the actual file
			file, err := files[i].Open()
			defer file.Close()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			//create destination file making sure the path is writeable.
			//log.Println("images file  ", files[i].Filename)
			dst, err := os.Create(img_dir + files[i].Filename)
			defer dst.Close()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			//copy the uploaded file to the destination file
			if _, err := io.Copy(dst, file); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			//make the images contents
			imgTitle := files[i].Filename
			last := strings.LastIndex(imgTitle, ".")

			if last > 0 && last < len(imgTitle) {
				imgTitle = imgTitle[0:last]
			}

			log.Println("imgTitle", imgTitle)
			imgContents += "![" + imgTitle + "](" + img_dir + files[i].Filename + ")\n"
		}
		//append uploaded images to end of md content
		content += imgContents
		mdData := MarkdownData{}
		pageData := PageData{Title: "GeoNet News"}
		mdData.Title = title
		mdData.MdContent = content
		pageData.MarkDown = mdData
		//show edit page again
		renderTemplate(w, "edit", pageData)

	} else { //submit
		//save md content
		saveMdContent(title, content)
		pageData := getHtmlPageData(title, content)
		//save html preview
		saveHtmlContent(title, "preview", pageData)
		indexPage(w, r)
	}
}

func previewPage(w http.ResponseWriter, r *http.Request, title string, content string) {
	pageData := getHtmlPageData(title, content)
	//save html preview
	saveHtmlContent(title, "preview", pageData)
	//render
	renderTemplate(w, "edit", pageData)
}

func getHtmlPageData(title string, md string) *PageData {
	mdData := MarkdownData{}
	pageData := PageData{Title: "GeoNet News"}
	mdData.Title = title
	mdData.MdContent = md

	htmlRenderer := blackfriday.HtmlRenderer(htmlFlags, "", "")
	mdData.HtmlContent = template.HTML(blackfriday.Markdown([]byte(md), htmlRenderer, htmlExt))
	pageData.MarkDown = mdData

	return &pageData
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
func saveHtmlContent(title string, tmplName string, data interface{}) {
	fleName := getFileNameForTitle(title)
	f, err := os.Create(md_html_dir + fleName + ".html")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	tmpl, ok := templates[tmplName]
	if !ok {
		panic("The template preview does not exist.")
	}
	err = tmpl.ExecuteTemplate(w, "base", &data)
	if err != nil {
		panic(err)
	}
	w.Flush()
}

func CopyDir(source string, dest string) (err error) {

	// get properties of source dir
	sourceinfo, err := os.Stat(source)
	if err != nil {
		return err
	}

	// create dest dir
	err = os.MkdirAll(dest, sourceinfo.Mode())
	if err != nil {
		return err
	}

	directory, _ := os.Open(source)
	objects, err := directory.Readdir(-1)

	for _, obj := range objects {
		sourcefilepointer := source + "/" + obj.Name()
		destinationfilepointer := dest + "/" + obj.Name()

		// perform copy
		err = CopyFile(sourcefilepointer, destinationfilepointer)
		if err != nil {
			fmt.Println(err)
		}
	}
	return
}

func CopyFile(source string, dest string) (err error) {
	sourcefile, err := os.Open(source)
	if err != nil {
		return err
	}

	defer sourcefile.Close()

	destfile, err := os.Create(dest)
	if err != nil {
		return err
	}

	defer destfile.Close()

	_, err = io.Copy(destfile, sourcefile)
	if err == nil {
		sourceinfo, err := os.Stat(source)
		if err != nil {
			err = os.Chmod(dest, sourceinfo.Mode())
		}

	}
	return
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
