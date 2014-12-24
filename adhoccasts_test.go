package main

import (
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	BASE_TEST_URL = "http://mytest.com"
)

func assertNoError(err error, t *testing.T) {
	if err != nil {
		t.Fatal(err)
	}
}

func TestIsDirectory_DetectsADirectory(t *testing.T) {
	dir, err := ioutil.TempDir("", "xmlfile")
	assertNoError(err, t)
	isDir, err := isDirectory(dir)
	assertNoError(err, t)
	if !isDir {
		t.Fatal("Expected a directory")
	}
}

func TestIsDirectory_DetectsAFile(t *testing.T) {
	file, err := ioutil.TempFile("", "xmlfile")
	assertNoError(err, t)
	isDir, err := isDirectory(file.Name())
	assertNoError(err, t)
	if isDir {
		t.Fatal("Expected a file")
	}
}

func TestIsDirectory_DetectsAnError(t *testing.T) {
	_, err := isDirectory("sfdkasdfjkasfdkjsakkdsffd dsffsdkj")
	if err == nil {
		t.Fatal("Expected an error")
	}
}

type Dir struct {
	Name        string
	Files       []File
	Directories []Dir
}

type File struct {
	Name string
}

func buildFSStructure(dir Dir, t *testing.T) string {
	var buildDir func(string, Dir)

	buildDir = func(oldRoot string, dir Dir) {
		rootDir := oldRoot + "/" + dir.Name

		err := os.Mkdir(rootDir, os.ModeDir|os.ModePerm)
		assertNoError(err, t)

		for _, file := range dir.Files {
			file, err := os.Create(rootDir + "/" + file.Name)
			assertNoError(err, t)
			file.Close()
		}
		for _, directory := range dir.Directories {
			buildDir(rootDir, directory)
		}
	}

	root, err := ioutil.TempDir("", "podcast")
	assertNoError(err, t)

	buildDir(root, dir)

	return root + "/" + dir.Name + "/"
}

func executeQueryAgainst(d Dir, queryPath string, t *testing.T) ([]byte, int) {
	dir := buildFSStructure(d, t)
	//defer os.RemoveAll(dir)

	absDir := filepath.Clean(dir)
	ts := httptest.NewServer(http.HandlerFunc(createCastHandler(BASE_TEST_URL, absDir)))
	defer ts.Close()

	res, err := http.Get(ts.URL + "/" + queryPath)
	assertNoError(err, t)
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	assertNoError(err, t)

	return body, res.StatusCode
}

func TestGetMp3File(t *testing.T) {
	resBody, statusCode := executeQueryAgainst(
		Dir{"root", []File{File{"my.mp3"}}, nil},
		"my.mp3", t)

	if statusCode != http.StatusOK {
		t.Fatal("Expected a 200")
	}

	if string(resBody) != "" {
		t.Fatal("Mp3 file should be empty")
	}
}

func TestGetListingFromRoot(t *testing.T) {
	zero := "zero"
	one := "one"
	two := "two"
	resBody, statusCode := executeQueryAgainst(
		Dir{"root",
			[]File{
				File{zero},
			},
			[]Dir{
				Dir{one, nil, nil},
				Dir{two, nil, nil},
			},
		},
		"/", t)

	if statusCode != http.StatusOK {
		t.Fatal("Expected a 200")
	}

	if strings.Contains(string(resBody), zero) {
		t.Fatalf("Expected to not find the %v fil in the listing. Got %v", zero, string(resBody))
	}
	if !strings.Contains(string(resBody), one) {
		t.Fatalf("Expected to find the %v podcast in the listing. Got %v", one, string(resBody))
	}
	if !strings.Contains(string(resBody), two) {
		t.Fatalf("Expected to find the %v podcast in the listing. Got %v", two, string(resBody))
	}
}

func TestGet403_WhenTryingToGetParentDirectory(t *testing.T) {
	_, statusCode := executeQueryAgainst(Dir{"root", nil, nil}, "../../", t)
	if statusCode != http.StatusForbidden {
		t.Fatalf("Expected a %v got %v", http.StatusForbidden, statusCode)
	}
}

func TestGet404_WhenPodcastDirDoesntExist(t *testing.T) {
	resBody, statusCode := executeQueryAgainst(Dir{"fooDir", nil, nil}, "xxx.xml", t)

	if statusCode != http.StatusNotFound {
		t.Fatal("Expected a 404")
	}

	expected := "No such podcast."
	if string(resBody) != expected {
		t.Fatalf("Expected (%v) got (%v)", expected, string(resBody))
	}
}

func TestGet404_WhenPodcastDirIsAFile(t *testing.T) {
	resBody, statusCode := executeQueryAgainst(
		Dir{"root", []File{File{"bar"}}, nil}, "bar.xml", t)

	if statusCode != http.StatusUnauthorized {
		t.Fatal("Expected a 401")
	}

	expected := "Not a podcast."
	if string(resBody) != expected {
		t.Fatalf("Expected (%v) got (%v)", expected, string(resBody))
	}
}

func executeValidQueryAgainst(d Dir, queryPath string, t *testing.T) Rss {
	resBody, statusCode := executeQueryAgainst(
		d, queryPath, t)

	if statusCode != http.StatusOK {
		t.Fatal("Expected a 200")
	}

	rss := Rss{}
	err := xml.Unmarshal(resBody, &rss)
	assertNoError(err, t)
	return rss
}

func TestGet_ReturnsCorrectTitle(t *testing.T) {
	title := "bar"
	rss := executeValidQueryAgainst(
		Dir{"root", nil, []Dir{Dir{title, nil, nil}}}, title+".xml", t)

	if rss.Channel.Title != title {
		t.Fatalf("Expected title of %v got %v", title, rss.Channel.Title)
	}
}

func TestGet_ReturnsCorrectLongTitle(t *testing.T) {
	title := "bar_is_bar"
	expectedTitle := "bar is bar"
	rss := executeValidQueryAgainst(
		Dir{"root", nil, []Dir{Dir{title, nil, nil}}}, title+".xml", t)

	if rss.Channel.Title != expectedTitle {
		t.Fatalf("Expected title of %v got %v", title, rss.Channel.Title)
	}
}

func TestGet_ReturnsCorrectDescription(t *testing.T) {
	title := "bar__baz"
	description := "baz"
	rss := executeValidQueryAgainst(
		Dir{"root", nil, []Dir{Dir{title, nil, nil}}}, title+".xml", t)

	if rss.Channel.Description != description {
		t.Fatalf("Expected title of %v got %v", title, rss.Channel.Title)
	}
}

func TestGet_ReturnsCorrectLongDescription(t *testing.T) {
	title := "bar__baz_is_a_baz"
	description := "baz is a baz"
	rss := executeValidQueryAgainst(
		Dir{"root", nil, []Dir{Dir{title, nil, nil}}}, title+".xml", t)

	if rss.Channel.Description != description {
		t.Fatalf("Expected title of %v got %v", title, rss.Channel.Title)
	}
}

func TestGet_ReturnsItemWithTitleFromMp3Name(t *testing.T) {
	title := "some__podcast"
	mp3Name := "my_name_is.mp3"
	mp3Title := "my name is"
	rss := executeValidQueryAgainst(
		Dir{"root", nil, []Dir{Dir{title, []File{File{mp3Name}}, nil}}}, title+".xml", t)

	if rss.Channel.Items[0].Title != mp3Title {
		t.Fatalf("Expected an item titled %v got %v", mp3Title, rss.Channel.Items[0].Title)
	}
}

func TestGet_ReturnsItemWithDescriptionFromMp3Name(t *testing.T) {
	title := "some__podcast"
	mp3Name := "my_name_is.mp3"
	mp3Description := "my name is"
	rss := executeValidQueryAgainst(
		Dir{"root", nil, []Dir{Dir{title, []File{File{mp3Name}}, nil}}}, title+".xml", t)

	if rss.Channel.Items[0].Description != mp3Description {
		t.Fatalf("Expected an item titled %v got %v", mp3Description, rss.Channel.Items[0].Description)
	}
}

func TestGet_ReturnsItemWithEnclosureFromMp3Name(t *testing.T) {
	title := "some__podcast"
	mp3Name := "my_name_is.mp3"
	enclosure := Enclosure{BASE_TEST_URL + "/" + title + "/" + mp3Name, 0, "audio/mpeg"}
	rss := executeValidQueryAgainst(
		Dir{"root", nil, []Dir{Dir{title, []File{File{mp3Name}}, nil}}}, title+".xml", t)

	if rss.Channel.Items[0].Enclosure != enclosure {
		t.Fatalf("Expected an enclosure %v got %v", enclosure, rss.Channel.Items[0].Enclosure)
	}
}

func TestGet_ReturnsItemWithGuidFromMp3Name(t *testing.T) {
	title := "some__podcast"
	mp3Name := "my_name_is.mp3"
	guid := Guid{true, BASE_TEST_URL + "/" + title + "/" + mp3Name}
	rss := executeValidQueryAgainst(
		Dir{"root", nil, []Dir{Dir{title, []File{File{mp3Name}}, nil}}}, title+".xml", t)

	if rss.Channel.Items[0].Guid != guid {
		t.Fatalf("Expected a guid %v got %v", guid, rss.Channel.Items[0].Guid)
	}
}

func TestGet_ReturnsOneItemPerFile(t *testing.T) {
	title := "some__podcast"
	count := 2
	rss := executeValidQueryAgainst(
		Dir{"root", nil, []Dir{Dir{title, []File{File{"a.mp3"}, File{"b.mp3"}}, nil}}}, title+".xml", t)

	if len(rss.Channel.Items) != 2 {
		t.Fatalf("Expected %v items got %v", count, len(rss.Channel.Items))
	}
}

func TestValidatePath_Handles_GoodDirs(t *testing.T) {
	if path, _ := validatePath("/"); path != "/" {
		t.Fatalf("Expecting a valid dir")
	}
}

func TestValidatePath_Handles_GoodFiles(t *testing.T) {
	if path, _ := validatePath("/dev/null"); path != "/dev/null" {
		t.Fatalf("Expecting a valid file")
	}
}

func TestValidateDir_HandlesGoodDirectory(t *testing.T) {
	if dir, _ := validateDir("/"); dir != "/" {
		t.Fatalf("Expecting a valid directory")
	}
}

func TestValidateDir_CleansDirectory(t *testing.T) {
	if dir, _ := validateDir("/./"); dir != "/" {
		t.Fatalf("Expecting a valid directory")
	}
}

func TestValidateDir_HandlesBadDirectory(t *testing.T) {
	if _, err := validateDir("sdfa"); err == nil {
		t.Fatalf("Expecting an error got valid directory")
	}
}

func TestGetAddress_HandlesInvalidAddress(t *testing.T) {
	if _, err := getAddress("::::sdfa"); err == nil {
		t.Fatalf("Expecting an error got an address")
	}
}

func TestGetAddress_HandlesAddressWithoutPort(t *testing.T) {
	if addr, _ := getAddress("http://foo.com"); addr != ":8080" {
		t.Fatalf("Expecting the default address")
	}
}

func TestGetAddress_HandlesAddressWithPort(t *testing.T) {
	if addr, _ := getAddress("http://foo.com:9999"); addr != ":9999" {
		t.Fatalf("Expecting the given address")
	}
}
