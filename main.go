package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

const (
	version       = "3.0.4"
	base          = "https://api.consumet.org/movies/flixhq"
	configFile    = "/.config/lobster/lobster_config.txt"
	historyFile   = "/.config/lobster/lobster_history.txt"
	playerDefault = "mpv"
	subsLanguage  = "English"
	videoQuality  = "1080p"
	serverDefault = "vidcloud"
)

func main() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	configFile := homeDir + configFile
	historyFile := homeDir + historyFile

	if _, err := os.Stat(homeDir + "/.config/lobster"); os.IsNotExist(err) {
		err = os.MkdirAll(homeDir+"/.config/lobster", os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
	}

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		f, err := os.Create(configFile)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		_, err = f.WriteString("player=mpv\nsubs_language=English\nvideo_quality=1080\npreferred_server=vidcloud\n")
		if err != nil {
			log.Fatal(err)
		}
	}

	player, subsLanguage, videoQuality, server := getConfig(configFile)

	separator, pathThing := getSeparator()

	playVideo(player, subsLanguage, videoQuality, server, separator, pathThing)
}

func getConfig(configFile string) (string, string, string, string) {
	file, err := os.Open(configFile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	player := playerDefault
	subsLanguage := subsLanguage
	videoQuality := videoQuality
	server := serverDefault
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "player=") {
	player = strings.TrimPrefix(line, "player=")
} else if strings.HasPrefix(line, "subs_language=") {
	subsLanguage = strings.TrimPrefix(line, "subs_language=")
} else if strings.HasPrefix(line, "video_quality=") {
	videoQuality = strings.TrimPrefix(line, "video_quality=")
} else if strings.HasPrefix(line, "preferred_server=") {
	server = strings.TrimPrefix(line, "preferred_server=")
}
	}

	return player, subsLanguage, videoQuality, server
}

func getSeparator() (string, string) {
	separator := ":"
	pathThing := "\\"
	if runtime.GOOS == "windows" {
		separator = ";"
		pathThing = ""
	}
	return separator, pathThing
}

func parseJSONData(jsonData, videoQuality, subsLanguage string) (string, string, string, string, error) {
	var referrer, mpvLink, subsLinks, videoTitle string
	lines := strings.Split(jsonData, "\n")
	for _, line := range lines {
		if strings.Contains(line, "\"Referer\":\"") {
			referrer = strings.TrimPrefix(line, "\"Referer\":\"")
			referrer = strings.TrimSuffix(referrer, "\"")
		} else if strings.Contains(line, "\"url\":\"") && strings.Contains(line, "\"quality\":\""+videoQuality+"\",") {
			mpvLink = strings.TrimPrefix(line, "\"url\":\"")
			mpvLink = strings.TrimSuffix(mpvLink, "\",\"quality\":\""+videoQuality+"\",")
		} else if strings.Contains(line, "\"url\":\"") && strings.Contains(line, "\"lang\":\""+subsLanguage+"\"") {
			subsLink := strings.TrimPrefix(line, "\"url\":\"")
			subsLink = strings.TrimSuffix(subsLink, "\",\"lang\":\""+subsLanguage+"\"")
			subsLinks += strings.Replace(subsLink, ":", pathThing+":", -1) + separator
		} else if strings.Contains(line, "\"title\":\"") {
			videoTitle = strings.TrimPrefix(line, "\"title\":\"")
			videoTitle = strings.TrimSuffix(videoTitle, "\"")
		}
	}
}

func playVideo(player, subsLanguage, videoQuality, server, separator, pathThing string) {
	episodeID := "12345"
mediaID := "67890"

jsonData, err := getJSONData(base + "/watch?episodeId=" + episodeID + "&mediaId=" + mediaID + "&server=" + server + "&")
if err != nil {
	log.Fatal(err)
}


	referrer, mpvLink, subsLinks, videoTitle, err := parseJSONData(jsonData, videoQuality, subsLanguage)
	if err != nil {
		log.Fatal(err)
	}

	if player == "iina" {
		cmd := exec.Command("iina", "--no-stdin", "--keep-running", "--mpv-referrer="+referrer, "--mpv-sub-files="+subsLinks, "--mpv-force-media-title="+videoTitle, mpvLink)
		err = cmd.Run()
		if err != nil {
			log.Fatal(err)
		}
	} else if player == "vlc" {
		if runtime.GOOS == "android" {
			cmd := exec.Command("am", "start", "--user", "0", "-a", "android.intent.action.VIEW", "-d", mpvLink, "-n", "org.videolan.vlc/org.videolan.vlc.gui.video.VideoPlayerActivity", "-e", "title", videoTitle)
			err = cmd.Run()
			if err != nil {
				log.Fatal(err)
			}
		} else {
			cmd := exec.Command("vlc", mpvLink, "--http-referrer="+referrer, "--meta-title", videoTitle)
			err = cmd.Run()
			if err != nil {
				log.Fatal(err)
			}
		}
	} else {
		if runtime.GOOS == "android" {
			cmd := exec.Command("am", "start", "--user", "0", "-a", "android.intent.action.VIEW", "-d", mpvLink, "-n", "is.tinyplayer/.TinyPlayerActivity", "-e", "title", videoTitle)
			err = cmd.Run()
			if err != nil {
				log.Fatal(err)
			}
		} else {
        cmd := exec.Command(player, "--http-referrer="+referrer, "--sub-files="+subsLinks, mpvLink)
			err = cmd.Run()
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}


func getJSONData(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}


