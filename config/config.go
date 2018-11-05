package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/davecgh/go-spew/spew"
	"github.com/fatih/color"
	"github.com/nmrshll/go-cp"
	"github.com/nmrshll/google-photos-api-client-go/lib-gphotos"
	"github.com/nmrshll/gphotos-uploader-cli/fileshandling"
	"github.com/palantir/stacktrace"
	"golang.org/x/oauth2"

	"github.com/client9/xson/hjson"
)

const (
	CONFIGFOLDER = "~/.config/gphotos-uploader-cli"
)

type APIAppCredentials struct {
	ClientID     string
	ClientSecret string
}

var (
	// consts
	CONFIGPATH                  = fmt.Sprintf("%s/config.hjson", CONFIGFOLDER)
	DEFAULT_API_APP_CREDENTIALS = APIAppCredentials{
		ClientID:     "20637643488-1hvg8ev08r4tc16ca7j9oj3686lcf0el.apps.googleusercontent.com",
		ClientSecret: "0JyfLYw0kyDcJO-pGg5-rW_P",
	}

	// vars
	Cfg *Config
)

type Config struct {
	APIAppCredentials *APIAppCredentials
	Jobs              []FolderUploadJob
}

func (c *Config) Process() {
	if c.APIAppCredentials == nil {
		c.APIAppCredentials = &DEFAULT_API_APP_CREDENTIALS
	}
}

func OAuthConfig() *oauth2.Config {
	if Cfg.APIAppCredentials == nil {
		log.Fatal(stacktrace.NewError("APIAppCredentials can't be nil"))
	}
	return gphotos.NewOAuthConfig(gphotos.APIAppCredentials(*Cfg.APIAppCredentials))
}

type FolderUploadJob struct {
	Account      string
	SourceFolder string
	MakeAlbums   struct {
		Enabled bool
		Use     string
	}
	DeleteAfterUpload bool
}

func Load() *Config {
	Cfg = loadConfigFile()
	Cfg.Process()
	return Cfg
}

func loadConfigFile() *Config {
	configPathAbsolute, err := cp.AbsolutePath(CONFIGPATH)
	if err != nil {
		log.Fatal(err)
	}

	// if no config file copy the default example and exit
	if !fileshandling.IsFile(configPathAbsolute) {
		fmt.Println(color.CyanString(`
No config file found at ~/.config/gphotos-uploader-cli/config.hjson
Will now copy an example config file.
Edit it by running:

	nano ~/.config/gphotos-uploader-cli/config.hjson

`,
		))
		spew.Dump(configPathAbsolute)
		dir := filepath.Dir(configPathAbsolute)
		err = os.MkdirAll(dir, 755)
		if err != nil {
			log.Fatal(err)
		}
		f, err := os.Create(configPathAbsolute)
		if err != nil {
			log.Fatal(err)
		}
		_, err = f.WriteString(ExampleConfig)
		if err != nil {
			log.Fatal(err)
		}

		os.Exit(0)
	}

	// else load and continue
	fmt.Println("[INFO] Config file found. Loading...")
	configBytes, err := ioutil.ReadFile(configPathAbsolute)
	if err != nil {
		log.Fatal(err)
	}
	var config = &Config{}
	jsonConfig := hjson.ToJSON(configBytes)
	if err := json.Unmarshal(jsonConfig, config); err != nil {
		log.Fatal(stacktrace.Propagate(err, "unmarshal jsonConfig failed"))
	}
	return config
}

const ExampleConfig = `{
  APIAppCredentials: {
    ClientID:     "20637643488-1hvg8ev08r4tc16ca7j9oj3686lcf0el.apps.googleusercontent.com",
    ClientSecret: "0JyfLYw0kyDcJO-pGg5-rW_P",
  }
  jobs: [
    {
      account: youremail@gmail.com
      sourceFolder: ~/folder/to/upload
      makeAlbums: {
        enabled: true
        use: folderNames
      }
      deleteAfterUpload: true
    }
  ]
}
`
