// Package generator generates the RSS feed.
package generator

import (
	"bytes"
	"fmt"
	htmltemplate "html/template"
	"log"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/BurntSushi/toml"
)

var CurrentConfig ConfigType
var CurrentFeedBytes []byte

type ConfigType struct {
	GeneralConfig GeneralConfigType
	Episode       []EpisodeConfigType
}

type EpisodeConfigType struct {
	EpisodeTitle       string
	EpisodeDescription string
	Audio              string
	PubDate            string
	ITunesSeason       string
	ITunesEpisode      string
	GeneratedValues    struct {
		EpisodeLink     string
		EpisodeFileSize string
		EpisodeType     string
		EpisodeDuration string

		DescriptionTrusted htmltemplate.HTML
		PubDateReadable    string
	} `toml:"_"`
}

type GeneralConfigType struct {
	URL               string
	Method            string
	Title             string
	Description       string
	Link              string
	ImageTitle        string
	Copyright         string
	Language          string
	ITunesAuthor      string
	ITunesType        string
	ITunesAuthorName  string
	ITunesAuthorEmail string
	ITunesExplicit    string
	ITunesCategory    string
	ITunesSubcategory string
	GeneratedValues   struct {
		ImageURL string
	} `toml:"_"`
}

func init() {
	conf, err := loadConfigFromFile()
	if err != nil {
		log.Fatal(err)
	}
	CurrentConfig = conf
	err = GenerateFeed(conf)
	if err != nil {
		log.Fatal(err)
	}
}

func loadConfigFromFile() (ConfigType, error) {
	// Read in configuration file
	content, err := readFile("config.toml")
	if err != nil {
		return ConfigType{}, err
	}

	var conf ConfigType
	if _, err := toml.Decode(string(content), &conf); err != nil {
		return ConfigType{}, err
	}
	return conf, nil
}

// GenerateFeed generates the podcast feed rss
func GenerateFeed(conf ConfigType) error {
	var generalConfig = &conf.GeneralConfig
	var episodes = conf.Episode

	generalConfig.GeneratedValues.ImageURL = generalConfig.Method + "://" + path.Join(generalConfig.URL, "download", "image.jpg")

	// Set automatic parameters
	for i, e := range episodes {
		e.GeneratedValues.EpisodeLink = generalConfig.Method + "://" + path.Join(generalConfig.URL, "download", "audio", e.Audio)

		audioFile := path.Join("static/download/audio/", e.Audio)

		// Determine file size of audio file
		size, err := fileSize(audioFile)
		if err != nil {
			return err
		}

		e.GeneratedValues.EpisodeFileSize = strconv.Itoa(int(size))

		// REALLY BAD HACK: Determining mime type

		splitFile := strings.Split(audioFile, ".")
		var mimeType string
		switch splitFile[len(splitFile)-1] {
		case "mp3":
			mimeType = "audio/mpeg"
		case "wav":
			mimeType = "audio/vnd.wav"
		case "ogg":
			mimeType = "audio/ogg"
		default:
			mimeType = "?"
		}
		e.GeneratedValues.EpisodeType = mimeType

		// REALLY REALLY BAD HACK: Determining epsiode length

		args := fmt.Sprintf("-i %s -show_entries format=duration -v quiet -of csv=p=0", audioFile)
		cmd := exec.Command("ffprobe", strings.Split(args, " ")...)
		outBytes, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("%v, do you have ffmpeg installed?", err)
		}

		durationFloat, err := strconv.ParseFloat(strings.Replace(string(outBytes), "\n", "", 1), 64)
		if err != nil {
			return err
		}
		e.GeneratedValues.EpisodeDuration = strconv.Itoa(int(durationFloat))

		pubDateTime, err := time.Parse("2 Jan 2006", e.PubDate)
		if err != nil {
			return err
		}
		e.GeneratedValues.PubDateReadable = pubDateTime.Format("2. Jan 2006")
		e.GeneratedValues.DescriptionTrusted = htmltemplate.HTML(e.EpisodeDescription)

		episodes[i] = e
	}

	var buffer bytes.Buffer

	rssTemplate, err := template.ParseFiles("templates/template.xml")
	if err != nil {
		return err
	}

	err = rssTemplate.Execute(&buffer, conf)
	if err != nil {
		return err
	}

	CurrentFeedBytes = buffer.Bytes()
	return nil
}

func readFile(name string) ([]byte, error) {
	file, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	fileinfo, err := file.Stat()
	if err != nil {
		return nil, err
	}
	buffer := make([]byte, fileinfo.Size())
	_, err = file.Read(buffer)
	if err != nil {
		return nil, err
	}
	return buffer, nil
}

func fileSize(name string) (int64, error) {
	file, err := os.Open(name)
	if err != nil {
		return 0, err
	}
	defer file.Close()
	fileinfo, err := file.Stat()
	if err != nil {
		return 0, err
	}
	return fileinfo.Size(), nil
}
