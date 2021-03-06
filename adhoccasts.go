package main

import (
	"encoding/xml"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

const (
	FORMAT  = "Mon Jan 2 15:04:05 -0700 MST 2006"
	LISTING = `<html>
        <head><title>Adhoc Podcasts</title></head>
        <body>
            {{range $name, $url := .NameToPath}}
                <a href="{{ $url }}">{{ $name }}</a><br>
            {{ end }}
        </body>
    </html>
    `
)

type podcasts struct {
	NameToPath map[string]string
}

func isDirectory(dir string) (bool, error) {
	stat, err := os.Stat(dir)
	if err != nil {
		return false, err
	}
	if stat.IsDir() {
		return true, nil
	}
	return false, nil
}

type Rss struct {
	XMLName xml.Name `xml:"rss"`
	Version string   `xml:"version,attr"`
	Channel Channel  `xml:"channel"`
}

type Channel struct {
	Title         string `xml:"title"`
	Description   string `xml:"description"`
	LastBuildDate string `xml:"lastBuildDate"`
	Items         []Item `xml:"item"`
}

type Item struct {
	Title       string    `xml:"title"`
	Description string    `xml:"description"`
	Enclosure   Enclosure `xml:"enclosure"`
	Guid        Guid      `xml:"guid"`
}

type Enclosure struct {
	Url    string `xml:"url,attr"`
	Length int64  `xml:"length,attr"`
	Type   string `xml:"type,attr"`
}

type Guid struct {
	IsPermaLink bool   `xml:"isPermaLink,attr"`
	Value       string `xml:"guid"`
}

func dirToTitle(dir string) string {
	title__description := strings.Split(filepath.Base(dir), "__")
	title := title__description[0]
	return strings.Replace(title, "_", " ", -1)
}

func dirToDescription(dir string) string {
	title__description := strings.Split(filepath.Base(dir), "__")
	description := ""
	if len(title__description) > 1 {
		description = title__description[1]
	}
	return strings.Replace(description, "_", " ", -1)
}

func createCastHandler(baseUrl string, rootDir string) func(http.ResponseWriter, *http.Request) {
	theHandler := func(w http.ResponseWriter, r *http.Request) {
		// Get the full clean path
		cleanPath, err := validatePath(path.Join(rootDir, r.URL.Path))
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(fmt.Sprintf("Bad path (%s) (%s).", r.URL.Path, err)))
			return
		}

		// Make sure user isn't reading outside of root directory
		if !strings.HasPrefix(cleanPath, rootDir) {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(fmt.Sprintf("Bad path (%s).", r.URL.Path)))
			return
		}

		// The podcast listing
		if cleanPath == rootDir {
			fileInfos, err := ioutil.ReadDir(rootDir)
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte("No such root directory."))
				return
			}

			p := podcasts{make(map[string]string)}
			for _, fileInfo := range fileInfos {
				if fileInfo.IsDir() {
					p.NameToPath[dirToTitle(fileInfo.Name())] = baseUrl + "/" + fileInfo.Name() + ".xml"
				}
			}
			templ, err := template.New("Listing").Parse(LISTING)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Server is broken."))
				return
			}
			templ.Execute(w, &p)
		} else if strings.HasSuffix(cleanPath, ".xml") {
			fmt.Printf("Generating %s\n", cleanPath)
			podcastDir := strings.TrimSuffix(cleanPath, filepath.Ext(cleanPath))
			isDir, err := isDirectory(podcastDir)
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte("No such podcast."))
				fmt.Println("Not Found")
				return
			} else if !isDir {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("Not a podcast."))
				fmt.Println("Not Authorized")
				return
			} else {
				title := dirToTitle(podcastDir)
				description := dirToDescription(podcastDir)

				fileInfos, err := ioutil.ReadDir(podcastDir)
				if err != nil {
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte("Can't read directory."))
					fmt.Println("Not Authorized (bad directory)")
					return
				}

				rss := Rss{
					xml.Name{"", ""},
					"2.0",
					Channel{title, description, time.Now().Format(FORMAT), nil},
				}

				items := []Item{}
				for _, fileInfo := range fileInfos {
					name := strings.TrimSuffix(fileInfo.Name(), filepath.Ext(fileInfo.Name()))
					name = strings.Replace(name, "_", " ", -1)
					url := baseUrl + "/" + path.Base(podcastDir) + "/" + fileInfo.Name()
					items = append(items, Item{
						name,
						name,
						Enclosure{
							url,
							fileInfo.Size(),
							"audio/mpeg",
						},
						Guid{
							true,
							url,
						}})
				}
				rss.Channel.Items = items
				encoder := xml.NewEncoder(w)
				encoder.Encode(rss)
				return
			}
		} else {
			fmt.Printf("Serving up %s\n", cleanPath)
			// A podcast mp3 file
			http.ServeFile(w, r, cleanPath)
		}
	}
	return theHandler
}

func validatePath(path string) (string, error) {
	cleanPath := filepath.Clean(path)
	cleanPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return "", errors.New(fmt.Sprintf("%s is not a valid path.", path))
	}

	return cleanPath, nil
}

func validateDir(dir string) (string, error) {
	cleanDir, err := validatePath(dir)
	if err != nil {
		return "", err
	}
	isDir, err := isDirectory(cleanDir)
	if !isDir || err != nil {
		return "", errors.New(fmt.Sprintf("%s is not a directory.", dir))
	}

	return cleanDir, nil
}

func getAddress(urlStr string) (string, error) {

	url, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}
	address := ":8080"
	host_address := strings.Split(url.Host, ":")
	if len(host_address) > 1 {
		address = ":" + host_address[1]
	}

	return address, nil
}

func main() {
	urlStr := flag.String("url", "http://localhost:8080", "The base url for the podcasts")
	dir := flag.String("dir", "./", "The directory where the adhoc podcasts are stored")
	flag.Parse()

	cleanDir, err := validateDir(*dir)
	if err != nil {
		log.Fatal(err)
	}

	address, err := getAddress(*urlStr)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Publishing directories under %s as podcasts on %s\n", *dir, *urlStr)
	http.HandleFunc("/", createCastHandler(*urlStr, cleanDir))
	log.Fatal(http.ListenAndServe(address, nil))
}
