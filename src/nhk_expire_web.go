/*
 * Copyright (c) 2025 metalfoxdev
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT, IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
*/

package main

import (
	"log"
	"flag"
	"io"
	"os"
	"net/http"
	"encoding/json"
	"net/url"
	"time"
	"strings"
	"fmt"
)

// Declare constants
const API_ROOT = "https://api.nhkworld.jp/"
const NW_ROOT = "https://www3.nhk.or.jp/"

// Initialise structs for video_episodes
type Video struct {
	ExpiredAt string `json:"expired_at"`
}
type VideoProg struct {
	Title string `json:"title"`
	Url string `json:"url"`
}
type Tag struct {
	Id string `json:"id"`
	Name string `json:"name"`
}
type Episode struct {
	Id string `json:"id"`
	Url string `json:"url"`
	Title string `json:"title"`
	Video Video `json:"video"`
	VideoProg VideoProg `json:"video_program"`
	Tags []Tag `json:"tags"`
}

type Pagination struct {
	Next string `json:"next"`
	Previous string `json:"previous"`
}

type Vods struct {
	Pagination Pagination `json:"pagination"`
	Episodes []Episode `json:"items"`
}

// Initialise structs for categories
type Category struct {
	Id string `json:"id"`
	Name string `json:"name"`
}
type Categories struct {
	Categories []Category `json:"items"`
}

// Init structs for JSON output
type OutVod struct {
	ProgName string `json:"prog_name"`
	EpName string `json:"ep_name"`
	ExpiredAt string `json:"expired_at"`
	ProgUrl string `json:"prog_url"`
	EpUrl string `json:"ep_url"`
}

type OutVods struct {
	LastUpdated string `json:"last_updated"`
	OutVod []OutVod `json:"progs"`
}

func getContent(url string) (content []byte, err error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

// Handles pagination
func getCatVideos(id string) (Vods, error) {
	var v Vods
	var e []Episode
	opt, _ := url.JoinPath("showsapi/v1/en/categories/", id, "/video_episodes")
	for i := 0; i < 50; i++ {
		v = Vods{}
		log.Printf("Reading page %d", i+1)
		path, _ := url.JoinPath(API_ROOT, opt)
		path = strings.Replace(path, "%3F", "?", -1) // Hacky solution lol
		resp, err := getContent(path)
		if err != nil {
			log.Fatalln(err)
		}
		json.Unmarshal(resp, &v)
		e = append(e, v.Episodes...)
		if v.Pagination.Next != "" {
			opt = v.Pagination.Next
		} else {
			v.Episodes = e
			return v, nil
		}
	}
	return v, fmt.Errorf("Pagination hit 50 page limit. Loop stopped to prevent infinite loop.")
}

// Dedupe function for the struct system
func existsInOutVods(ov OutVod, ovs []OutVod) (bool) {
	for i := 0; i < len(ovs); i++ {
		if ov == ovs[i] {
			return true
		}
	}
	return false
}

// Special check for Documentary 360
func checkD360(t []Tag) (bool) {
	for i := 0; i < len(t); i++ {
		if (t[i].Id == "196" && t[i].Name == "Documentary 360") || (t[i].Name == "Documentary 360") {
			return true
		}
	}
	return false
}

func main() {

	// Parse cmd flags
	filePtr := flag.String("file", "./progs.json", "Output file for result")
	flag.Parse()

	// Set offset for max threshold, 120 hours = 5 days
	expiryOffset, _ := time.ParseDuration("120h")

	// Get category list
	catListUrl, _ := url.JoinPath(API_ROOT, "/showsapi/v1/en/categories/")
	resp, err := getContent(catListUrl)
	if err != nil {
		log.Fatalln(err)
	}
	var categories Categories
	json.Unmarshal(resp, &categories)
	log.Printf("Loaded %d categories", len(categories.Categories))

	// Init variables for next step
	var vods Vods
	var ovs OutVods
	var ov OutVod
	var episodeUrl string
	var progUrl string
	timeNow := time.Now()
	expiryMaxDate := timeNow.Add(expiryOffset)

	// Log expiry max date
	log.Printf("Max expiry date set to '%s'", expiryMaxDate)

	// Process categories
	for i := 0; i < len(categories.Categories); i++ {
		vods = Vods{}
		log.Printf("Scanning category '%s'", categories.Categories[i].Name)

		// Get category videos
		v, err := getCatVideos(categories.Categories[i].Id)
		if err != nil {
			log.Fatalln(err)
		}

		// Declare vods
		vods = v

		// Scan category for expiring programmes
		for j := 0; j < len(vods.Episodes); j++ {

			// Convert ISO date to time struct
			expiryDate, err := time.Parse(time.RFC3339, vods.Episodes[j].Video.ExpiredAt)
			if err != nil {
				log.Fatalln(err)
			}

			// Check if expiryDate meets threshold
			// If so, add it to an array
			if expiryMaxDate.Unix() > expiryDate.Unix() {
				episodeUrl, _ = url.JoinPath(NW_ROOT, vods.Episodes[j].Url)
				progUrl, _ = url.JoinPath(NW_ROOT, vods.Episodes[j].VideoProg.Url)

				// Set programme attributes
				ov.ProgName = vods.Episodes[j].VideoProg.Title
				if checkD360(vods.Episodes[j].Tags) {
					// Documentary 360's aren't grouped
					// by show, they are grouped by tag.
					// Therefore, special operations are conducted.
					ov.ProgName = "Documentary 360"
					ov.EpName = vods.Episodes[j].VideoProg.Title
				} else if vods.Episodes[j].Title == "" {
					// Handle single progs
					// that are not part of a registered show
					ov.EpName = ov.ProgName
				} else {
					ov.EpName = vods.Episodes[j].Title
				}
				ov.ExpiredAt = vods.Episodes[j].Video.ExpiredAt
				ov.ProgUrl = progUrl
				ov.EpUrl = episodeUrl

				// Debuggering stuff
				//log.Printf("Prog Name : %s", ov.ProgName)
				//log.Printf("Ep Name   : %s", ov.EpName)

				// Append if not already listed
				if !existsInOutVods(ov, ovs.OutVod) {
					ovs.OutVod = append(ovs.OutVod, ov)
				}
			}
		}
	}

	// Report no. of matches found
	log.Printf("Found %d matching VODs", len(ovs.OutVod))
	ovs.LastUpdated = time.Now().Format(time.UnixDate)

	// Write matches to text file
	f, err := os.Create(*filePtr)
	defer f.Close()
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("Writing to %s...", *filePtr)
	ovJson, err := json.Marshal(ovs)
	if err != nil {
		log.Fatalln(err)
	}
	f.Write(ovJson)
}
