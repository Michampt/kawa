package main

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var httpClient = &http.Client{Timeout: 1 * time.Second}
var verbose bool

type ManifestItem struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type Manifest struct {
	Modules []ManifestItem `json:"modules"`
}

type Module struct {
	Name        string     `json:"name"`
	Version     string     `json:"version"`
	Description string     `json:"description"`
	Author      string     `json:"author"`
	Repo        string     `json:"repo"`
	Files       []FileInfo `json:"files"`
}

type FileInfo struct {
	Name string `json:"filename"`
	Hash string `json:"hash"`
}

func check(e error, message string) {
	if e != nil {
		fmt.Fprintf(os.Stderr, e.Error()+"\n"+message+"\n")
		os.Exit(1)
	}
}

func listModules() {
	manifest := &Manifest{}

	err := getJsonResponse("http://localhost:2209/.manifests/manifest.json", manifest)
	check(err, "Could not find manifest")

	fmt.Println("")
	fmt.Println("Name					Version")
	fmt.Println("----					-------")
	fmt.Println("")
	for _, module := range manifest.Modules {
		fmt.Println(module.Name + "				" + module.Version)
	}
}

func getModuleInfo(moduleName string) {
	module := &Module{}

	err := getJsonResponse(fmt.Sprintf("http://localhost:2209/.manifests/%s.json", moduleName), module)
	check(err, "Could not find specified module")

	fmt.Println("Name: " + module.Name)
	fmt.Println("Version: " + module.Version)
	fmt.Println("Description: " + module.Description)
	fmt.Println("Author: " + module.Author)
	fmt.Println("Repository: " + module.Repo)
	if verbose {
		fmt.Println("Files:")
		for _, file := range module.Files {
			fmt.Println(fmt.Sprintf("    %s    %s", file.Name, file.Hash))
		}
	}

}

func downloadModule(moduleName string) {
	response, err := http.Get(fmt.Sprintf("http://localhost:2209/%s.zip", moduleName))
	check(err, "Error downloading module")
	defer response.Body.Close()
	out, err := os.Create(moduleName + ".zip")
	check(err, "Error saving module to disk")
	defer out.Close()

	_, err = io.Copy(out, response.Body)
	check(err, "")

	_, err = unzipFile(moduleName+".zip", "app/modules/")
	check(err, "Error unzipping module")

	out.Close()

	err = os.Remove(moduleName + ".zip")
	check(err, "Error deleting module")
}

func unzipFile(src, dest string) ([]string, error) {
	var filenames []string

	zipReader, err := zip.OpenReader(src)
	check(err, "Error opening zip for reading")
	defer zipReader.Close()

	for _, zippedFile := range zipReader.File {
		fpath := filepath.Join(dest, zippedFile.Name)

		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return filenames, fmt.Errorf("%s: illegal file path", "fpath")
		}

		filenames = append(filenames, fpath)

		if zippedFile.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return filenames, err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, zippedFile.Mode())
		check(err, "Error opening file "+zippedFile.Name)

		rc, err := zippedFile.Open()
		check(err, "Error opening file "+zippedFile.Name)

		_, err = io.Copy(outFile, rc)
		check(err, "Error writing file "+zippedFile.Name)

		outFile.Close()
		rc.Close()
	}
	zipReader.Close()
	return filenames, nil
}

func getJsonResponse(url string, manifest interface{}) error {
	response, err := httpClient.Get(url)
	check(err, "")

	if response.StatusCode != 200 {
		return errors.New("Server responded with " + fmt.Sprint(response.StatusCode))
	}

	defer response.Body.Close()

	return json.NewDecoder(response.Body).Decode(manifest)
}

func removeModule(moduleName string) {
	err := os.RemoveAll(fmt.Sprintf("app/modules/%s", moduleName))
	check(err, "Error deleting module "+moduleName)
}

func init() {
	flag.BoolVar(&verbose, "v", false, "Verbose output")
	flag.Parse()
}

func main() {
	command := flag.Args()[0]
	switch command {
	case "list":
		listModules()
		break
	case "info":
		if len(flag.Args()) < 2 {
			fmt.Println("Missing module name")
			os.Exit(1)
		}
		getModuleInfo(flag.Args()[1])
		break
	case "install":
		if len(flag.Args()) < 2 {
			fmt.Println("Missing module name")
			os.Exit(1)
		}
		downloadModule(flag.Args()[1])
		break
	case "remove":
		if len(flag.Args()) < 2 {
			fmt.Println("Missing module name")
			os.Exit(1)
		}
		removeModule(flag.Args()[1])
	}
}
