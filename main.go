package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/mail"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"
)

// xml structs
type GoSortMail struct {
	XMLName xml.Name `xml:"gosortmail"`
	Default string   `xml:"default"`
	Logfile string   `xml:"logfile"`
	Rules   Rules    `xml:"rules"`
}

type Rules struct {
	XMLName xml.Name `xml:"rules"`
	Rule    []Rule   `xml:"rule"`
}

type Rule struct {
	Name     string `xml:"name"`
	Section  string `xml:"section"`
	Contains string `xml:"contains"`
	Folder   string `xml:"folder"`
}

// load config from /home/user/.gosortmailrc
func loadConfig(rulesFile string) GoSortMail {
	file, err := os.Open(rulesFile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	b, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}

	//fmt.Printf("%s", b)

	var gsm GoSortMail
	err = xml.Unmarshal(b, &gsm)
	if err != nil {
		log.Fatal(err)
	}

	return gsm
}

func fileExist(filePath string) bool {
	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false
	}
	if info.IsDir() {
		return false
	}
	return true
}

func createTmpFile(folderName string, logger *log.Logger) *os.File {
	now := time.Now()
	epoch := now.UnixNano()
	seconds := fmt.Sprintf("%v", epoch)

	tmpfile, err := ioutil.TempFile(folderName, seconds+".*.gosortmail")
	if err != nil {
		logger.Fatal(err)
	}
	return tmpfile
}

func makeMaildir(name string, logger *log.Logger) {
	err := os.MkdirAll(filepath.Join(name, "cur"), 0700)
	if err != nil {
		logger.Fatal(err)
	}

	err = os.MkdirAll(filepath.Join(name, "new"), 0700)
	if err != nil {
		logger.Fatal(err)
	}

	err = os.MkdirAll(filepath.Join(name, "tmp"), 0700)
	if err != nil {
		logger.Fatal(err)
	}
}

func deliverMail(folder string, msgBytes []byte, logger *log.Logger) {
	makeMaildir(folder, logger)

	tmpFolder := filepath.Join(folder, "tmp")
	tmpFile := createTmpFile(tmpFolder, logger)
	defer tmpFile.Close()

	if _, err := tmpFile.Write(msgBytes); err != nil {
		logger.Fatal(err)
	}

	tmpPath := tmpFile.Name()
	tmpFile.Close()

	base := filepath.Base(tmpPath)
	newPath := filepath.Join(folder, "new", base)

	if !fileExist(newPath) {
		os.Rename(tmpPath, newPath)
	}

	logger.Printf("%s is %d bytes\n", newPath, len(msgBytes))
}

func main() {
	// Get the user's home folder
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	// Load the config
	gsmConf := loadConfig(filepath.Join(usr.HomeDir, ".gosortmailrc"))

	// Open Logfile
	lf, err := os.OpenFile(gsmConf.Logfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		log.Println(err)
	}
	defer lf.Close()

	logger := log.New(lf, "gosortmail: ", log.LstdFlags)

	// Read the whole email message from stdin
	msgBytes, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		logger.Fatal(err)
	}

	// Parse the email header
	r := bytes.NewReader(msgBytes)
	m, err := mail.ReadMessage(r)
	if err != nil {
		logger.Fatal(err)
	}

	header := m.Header
	//fmt.Println("From:", header.Get("From"))
	//fmt.Println("To:", header.Get("To"))
	//fmt.Println("Cc:", header.Get("Cc"))
	//fmt.Println("Subject:", header.Get("Subject"))

	body, err := ioutil.ReadAll(m.Body)
	if err != nil {
		log.Fatal(err)
	}
	strBody := string(body)

	match := false

	for _, rule := range gsmConf.Rules.Rule {
		if strings.ToLower(rule.Section) == "body" {
			if strings.Contains(strBody, rule.Contains) {
				logger.Printf("Matched %s '%s'. Sorting to %s\n", rule.Name,
					header.Get("Subject"),
					rule.Folder)
				deliverMail(rule.Folder, msgBytes, logger)
				match = true
				break
			}
		} else {
			headerContent := header.Get(rule.Section)

			if strings.Contains(headerContent, rule.Contains) {
				logger.Printf("Matched %s '%s'. Sorting to %s\n", rule.Name,
					header.Get("Subject"),
					rule.Folder)
				deliverMail(rule.Folder, msgBytes, logger)
				match = true
				break
			}
		}
	}

	if !match {
		logger.Printf("No match '%s'. Sorting to %s\n", header.Get("Subject"), gsmConf.Default)
		deliverMail(gsmConf.Default, msgBytes, logger)
	}
}
