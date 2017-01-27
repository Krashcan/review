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
	"strconv"
//	"sync"
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

type node struct{
	Movie fileInfo 
	Left *(node) 
	Right *(node) 
}

var movie struct{
	Name string;
	Year string;
}
var root *node
var Movies []fileInfo
func main() {
//	goGroup := new(sync.WaitGroup)
	
	flag.Parse()
	files, _ := ioutil.ReadDir(flag.Args()[0])
	var queryNames []string
	for _, f := range files {
       	queryNames= append(queryNames,url.QueryEscape(f.Name()))
    }
	fmt.Println("Preparing data")
	
    for i:=0;i<len(queryNames);i++{
		go GetTitleAndYear("https://opensubtitles.co/search?q=" + queryNames[i])
    }
    
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

func GetTitleAndYear(url string){
	resp,err := http.Get(url)
	if err!=nil{
		fmt.Println(err)
		GetTitleAndYear(url)
		return
	}
	defer resp.Body.Close()
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
	x := fileInfo{}
	jsonParser := json.NewDecoder(req.Body)
    if err := jsonParser.Decode(&x); err != nil {
        log.Fatal("parsing config file", err)
    }
    if x == (fileInfo{}){
     	return
    }
    root = InsertTree(root,x)
    Movies = nil
    
    InorderTraversal(root)
    fmt.Println(x.Title,x.Year)
 }

func InsertTree(leaf *node,x fileInfo) *node{
	a,_ := strconv.ParseFloat(x.Rating,32)
	
	if leaf == nil{
		return &node{x,nil,nil}
	}else if b,_ := strconv.ParseFloat(leaf.Movie.Rating,32); a>b{
		leaf.Left = InsertTree(leaf.Left,x)
		return leaf
	}
	leaf.Right = InsertTree(leaf.Right,x)
	return leaf
	
}

func InorderTraversal(leaf *node){
	if leaf == nil{
		return
	}
	InorderTraversal(leaf.Left)
	Movies = append(Movies,leaf.Movie)
	InorderTraversal(leaf.Right)
}

