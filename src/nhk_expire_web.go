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
)

// Initialise structs for video_episodes
type Video struct {
	ExpiredAt string `json:"expired_at"`
}
type VideoProg struct {
	Title string `json:"title"`
	Url string `json:"url"`
}
type Episode struct {
	Id string `json:"id"`
	Url string `json:"url"`
	Title string `json:"title"`
	Video Video `json:"video"`
	VideoProg VideoProg `json:"video_program"`
}
type Vods struct {
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

func removeDuplicates(xs *[]string) {
	found := make(map[string]bool)
	j := 0
	for i, x := range *xs {
		if !found[x] {
			found[x] = true
			(*xs)[j] = (*xs)[i]
			j++
		}
	}
	*xs = (*xs)[:j]
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

func main() {

	// Parse cmd flags
	filePtr := flag.String("file", "./progs.json", "Output file for result")
	flag.Parse()

	// Declare constants
	const API_ROOT = "https://api.nhkworld.jp/showsapi/v1/"
	const NW_ROOT = "https://www3.nhk.or.jp/"

	// Set offset for max threshold, 120 hours = 5 days
	expiryOffset, _ := time.ParseDuration("120h")

	// Get category list
	catListUrl, _ := url.JoinPath(API_ROOT, "en/categories/")
	resp, err := getContent(catListUrl)
	if err != nil {
		log.Fatalln(err)
	}
	var categories Categories
	json.Unmarshal(resp, &categories)
	log.Printf("Loaded %d categories", len(categories.Categories))

	// Init variables for next step
	var catUrl string
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
		log.Printf("Scanning category '%s'", categories.Categories[i].Name)
		catUrl, _ = url.JoinPath(API_ROOT, "en/categories/", categories.Categories[i].Id, "/video_episodes")

		// Get category videos
		resp, err := getContent(catUrl)
		if err != nil {
			log.Fatalln(err)
		}
		json.Unmarshal(resp, &vods)

		// Scan category for expiring programmes
		for i2 := 0; i2 < len(vods.Episodes); i2++ {

			// Convert ISO date to time struct
			expiryDate, err := time.Parse(time.RFC3339, vods.Episodes[i2].Video.ExpiredAt)
			if err != nil {
				log.Fatalln(err)
			}

			// Check if expiryDate meets threshold
			// If so, add it to an array
			if expiryMaxDate.Unix() > expiryDate.Unix() {
				episodeUrl, _ = url.JoinPath(NW_ROOT, vods.Episodes[i2].Url)
				progUrl, _ = url.JoinPath(NW_ROOT, vods.Episodes[i2].VideoProg.Url)

				// Set programme attributes
				ov.ProgName = vods.Episodes[i2].VideoProg.Title
				ov.EpName = vods.Episodes[i2].Title
				ov.ExpiredAt = vods.Episodes[i2].Video.ExpiredAt
				ov.ProgUrl = progUrl
				ov.EpUrl = episodeUrl

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
