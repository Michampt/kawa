package main

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type zipCloser interface {
	Close() error
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

type ManifestItem struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type Manifest struct {
	Modules []ManifestItem `json:"modules"`
}

func closeZip(z zipCloser) {
	err := z.Close()
	check(err)
}

func readAll(file *zip.File) []byte {
	zc, err := file.Open()
	check(err)
	defer closeZip(zc)

	content, err := ioutil.ReadAll(zc)
	check(err)

	return content
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func createManifest() {
	fmt.Println("Recreating manifest directory")
	err := os.RemoveAll(".manifests")
	check(err)
	err = os.Mkdir(".manifests", 0755)
	check(err)

	manifestFile, err := os.Create(".manifests/manifest.json")
	manifest := &Manifest{Modules: []ManifestItem{}}

	var files []string

	root := "."
	fmt.Println("Scanning for modules")
	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && filepath.Ext(path) == ".zip" {
			files = append(files, path)
		}
		return nil
	})

	check(err)

	for _, zipFile := range files {
		// placeholder for module file that will be made
		var moduleFile *os.File

		// Module struct for details of the module

		openedZip, err := zip.OpenReader(zipFile)
		check(err)
		defer closeZip(openedZip)

		validModule, newModule, newManifestItem := checkManifest(openedZip)

		if !validModule {
			fmt.Println(zipFile + " is not a valid Mura module")
			continue
		}

		manifest.Modules = append(manifest.Modules, *newManifestItem)

		for _, zippedFile := range openedZip.File {
			if zippedFile.FileInfo().IsDir() {
				continue
			}

			zippedFileContents := readAll(zippedFile)
			fileName := zippedFile.FileInfo().Name()

			// get the files name and sum, add to the struct, then append to the newModule
			fileSha256 := sha256.Sum256(zippedFileContents)
			fileInfo := FileInfo{Name: fileName, Hash: hex.EncodeToString(fileSha256[:])}
			newModule.Files = append(newModule.Files, fileInfo)

			if fileName == "mura-module.json" {
				json.Unmarshal(zippedFileContents, &newModule)

				if validModule {
					// create the module file to be written
					moduleFile, err = os.Create(fmt.Sprintf(".manifests/%s.json", newModule.Name))
					check(err)

					// add the name and version to the main manifest item

				} else {
					continue
				}
			}
		}
		moduleJson, err := json.MarshalIndent(newModule, "", "    ")
		check(err)
		moduleFile.WriteString(string(moduleJson))
	}
	manifestJson, err := json.MarshalIndent(manifest, "", "    ")
	check(err)
	manifestFile.WriteString(string(manifestJson))
}

func checkManifest(zippedFile *zip.ReadCloser) (bool, *Module, *ManifestItem) {
	newModule := &Module{Files: []FileInfo{}}
	newManifestItem := &ManifestItem{}
	for _, file := range zippedFile.File {
		if file.FileInfo().Name() == "mura-module.json" {
			zippedFileContents := readAll(file)
			json.Unmarshal(zippedFileContents, &newModule)

			if newModule.Name == "" || newModule.Version == "" {
				return false, nil, nil
			}

			newManifestItem.Name = newModule.Name
			newManifestItem.Version = newModule.Version
			return true, newModule, newManifestItem
		}
	}
	return false, nil, nil
}

func startModuleServer(dir string) {
	var newDir string
	if strings.Contains(dir, "..") {
		currentDir, err := os.Getwd()
		check(err)

		newDir = filepath.Dir(filepath.Clean(filepath.Join(currentDir, dir)))
	} else {
		newDir = dir
	}
	fmt.Println("Starting http file server at " + newDir)
	fs := http.FileServer(http.Dir(dir))
	log.Fatal(http.ListenAndServe(":2209", fs))
}

func main() {
	currentDir, err := os.Getwd()
	check(err)

	serve := flag.Bool("s", false, "Whether to start the server or not")
	serverDirectory := flag.String("d", currentDir, "Directory to serve from")
	flag.Parse()

	switch *serve {
	case false:
		createManifest()
		break
	case true:
		startModuleServer(*serverDirectory)
		break
	}
}
