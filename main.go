package main

import(
	"net/http"
	"net/url"
	"log"
	"flag"
	"io/ioutil"
	"fmt"
	"golang.org/x/net/html"
	"strings"
	"encoding/json"
	"os"
	"html/template"
)

type fileInfo struct{
	Title string `json:"Title"`;
	Year string `json:"Year"`;
	Runtime string `json:"Runtime"`;
	Genre string `json:"Genre"` ;
	Rating string `json:"imdbRating"`;
	Description string `json:"Plot"`;
	Image string `json:"Poster"`;
	Awards string `json:"Awards"`;
}

var movie struct{
	Name string;
	Year string;
}
var Movies []fileInfo
func main() {
	flag.Parse()
	files, _ := ioutil.ReadDir(flag.Args()[0])
	var queryNames []string
	for _, f := range files {
       	queryNames= append(queryNames,url.QueryEscape(f.Name()))
    }
	//fmt.Println(os.Getenv("GOPATH") + "/src/github.com/krashcan/review/template/index.tpl")
	fmt.Println("Preparing data")
	
    for i,f:=range queryNames{
		go GetTitleAndYear("https://opensubtitles.co/search?q=" + f,&i)
    }

    fmt.Println("Preparation DONE")
	http.HandleFunc("/",ShowRatings)
    http.Handle("/static/",http.StripPrefix("/static/",http.FileServer(http.Dir(os.Getenv("GOPATH") + "/src/github.com/krashcan/review/static"))))
    
    log.Fatal(http.ListenAndServe(":8080",nil))
}

func ShowRatings(w http.ResponseWriter,r *http.Request){
	t,err := template.ParseFiles(os.Getenv("GOPATH") + "/src/github.com/krashcan/review/template/index.tpl")
	if(err!=nil){
		log.Fatal(err)
	}
	t.Execute(w,Movies)
}

func GetTitleAndYear(url string,i *int){
	resp,err := http.Get(url)
	if err!=nil{
		fmt.Println(err)
		(*i)--
		return
	}
	var movieData string
	if resp.StatusCode != 200 {
		fmt.Println("statuscode",err)
	}
	z := html.NewTokenizer(resp.Body)
	for{
		tt := z.Next()

		if tt == html.ErrorToken{
			return
		}else if tt==html.StartTagToken{
			t:= z.Token()
			if t.Data=="h4"{
				tt = z.Next()
				tt = z.Next()
				tt = z.Next()
				t = z.Token()
				movieData = strings.TrimSpace(t.Data)				
				break
			}
		}
	}

	movie.Name = movieData[:len(movieData)-6]
	movie.Year = movieData[len(movieData)-5:len(movieData)-1]
	movie.Name = strings.Replace(movie.Name, " ", "+", -1)
	url = "http://www.omdbapi.com/?t=" + movie.Name + "&y=" + movie.Year + "&plot=short&r=json"  
	req,err := http.Get(url)

	if err!=nil{
		log.Fatal(err)
	}
	var x fileInfo
	jsonParser := json.NewDecoder(req.Body)
    if err := jsonParser.Decode(&x); err != nil {
        log.Fatal("parsing config file", err)
    }
    Movies = append(Movies,x)
    fmt.Println(x.Title,x.Year)
}