package crawler

import (
	"encoding/json"
	"fmt"
	"log"
	"lsc/pkg/client"
	"lsc/pkg/utils"
	"os"
	"strings"
	"time"
)

const (
	TEMP_FILE_PATH = "./temp_problemset.txt"
	CONFIG_PATH    = "./config/config.json"
	LIMIT          = 40
	PAGE_TIME      = 3
	START_PAGE     = 0
)

type Config struct {
	Cookie    string `json:"cookie"`
	OutputDir string `json:"output_dir"`
	Day       int    `json:"day"`
	PushDir   string `json:"push_dir"`
}

type Args struct {
	Cookie    string
	Output    string
	Day       int
	Overwrite bool
}

type Crawler struct {
	Cookie                  string                 // cookie
	OutputDir               string                 // output directory
	TimeControl             int64                  // time control in seconds
	Overwrite               bool                   // whether to overwrite existing config
	PushDir                 string                 // directory to push to git
	C                       int                    // current page number
	Visited                 map[string]string      // map of visited problems
	ProblemsToBeReprocessed []string               // list of problems to be reprocessed
	Lc                      *client.LeetcodeClient // leetcode client
}

func NewCrawler(args Args) *Crawler {
	configFile, err := os.ReadFile(CONFIG_PATH)
	if err != nil {
		log.Fatal("Failed to read config file:", err)
	}

	var config Config
	if err := json.Unmarshal(configFile, &config); err != nil {
		log.Fatal("Failed to unmarshal config file:", err)
	}

	cookie := args.Cookie
	if cookie == "" {
		cookie = config.Cookie
	}

	outputDir := args.Output
	if outputDir == "" {
		outputDir = config.OutputDir
	}

	timeControl := int64(3600 * 24 * args.Day)
	if args.Day == 0 {
		timeControl = int64(3600 * 24 * config.Day)
	}

	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
			log.Fatal("Failed to create output directory:", err)
		}
	}

	lc := client.NewLeetcodeClient(cookie)

	return &Crawler{
		Cookie:                  cookie,
		OutputDir:               outputDir,
		TimeControl:             timeControl,
		Overwrite:               args.Overwrite,
		PushDir:                 config.PushDir,
		C:                       0,
		Visited:                 make(map[string]string),
		ProblemsToBeReprocessed: []string{},
		Lc:                      lc,
	}
}

func (c *Crawler) IsExpired(submission client.Submission) bool {
	curTime := time.Now().Unix()
	return curTime-submission.Timestamp > c.TimeControl
}

func (c *Crawler) ProcessSubmissions(submissions []client.Submission) bool {
	failCount := 0
	for _, submission := range submissions {
		if submission.StatusDisplay != "Accepted" {
			continue
		}
		if c.IsExpired(submission) {
			return true
		}
		if err := c.ProcessSubmission(submission); err != nil {
			log.Println("Error processing submission:", err)
			if failCount == 2 {
				log.Println("Service might be banning. Waiting before retrying...")
				break
			}
			failCount++
			time.Sleep(PAGE_TIME * 2 * time.Second)
		}
		time.Sleep(10 * time.Second)
	}
	return false
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		log.Fatal(err)
	}
	return true
}

func (c *Crawler) ProcessSubmission(submission client.Submission) error {
	submissionDetails, err := c.Lc.DownloadCode(submission)
	if err != nil {
		return err
	}

	problemFrontendID := submissionDetails.QuestionID
	problemTitle := submissionDetails.QuestionTitle
	submissionLang := submission.Lang
	submissionToken := problemTitle + submissionLang

	if _, visited := c.Visited[submissionToken]; !visited {
		c.Visited[submissionToken] = problemFrontendID
		fullPath, err := utils.GeneratePath(problemFrontendID, problemTitle, submissionLang, c.OutputDir)
		if err != nil {
			log.Println(err) //TODO
			return err
		}

		log.Println(fullPath)
		if !c.Overwrite && fileExists(fullPath) {
			return nil
		}
		if err := c.SaveCode(submissionDetails.Code, problemFrontendID, problemTitle, submissionLang, fullPath); err != nil {
			return err
		}
	}
	return nil
}

func (c *Crawler) SaveCode(code, problemFrontendID, problemTitle, submissionLang, fullPath string) error {
	if err := os.WriteFile(fullPath, []byte(code), 0644); err != nil {
		return fmt.Errorf("failed to write code to file: %w", err)
	}
	log.Println("Writing ends! " + fullPath)
	if c.IsTemporaryProblem(problemFrontendID) {
		c.ProblemsToBeReprocessed = append(c.ProblemsToBeReprocessed, fmt.Sprintf("%s %s %s", fullPath, problemTitle, submissionLang))
	}
	return nil
}

func (c *Crawler) IsTemporaryProblem(problemFrontendID string) bool {
	if len(problemFrontendID) > 0 && problemFrontendID[0] >= '6' {
		return true
	}
	return false
}

func (c *Crawler) ProcessTemporaryProblems() {
	if fileExists(TEMP_FILE_PATH) {
		tempFile, err := os.ReadFile(TEMP_FILE_PATH)
		if err != nil {
			log.Fatal("Failed to read temporary file:", err)
		}
		lines := strings.Split(string(tempFile), "\n")
		for _, line := range lines {
			if line == "" {
				continue
			}
			parts := strings.SplitN(line, " ", 2)
			if len(parts) != 3 {
				log.Println("Invalid format in temporary file, deleting all temp code.")
				os.Remove(parts[1])
				continue
			}
			path, title, lang := parts[0], parts[1], parts[2]
			token := title + lang
			if visitedProblemFrontendID, visited := c.Visited[token]; visited {
				if !c.IsTemporaryProblem(visitedProblemFrontendID) {
					log.Println(path + " is no longer a temporary problem, deleting temp code.")
					os.Remove(path)
				} else {
					c.ProblemsToBeReprocessed = append(c.ProblemsToBeReprocessed, fmt.Sprintf("%s %s %s", path, title, lang))
				}
			}
		}
	}
}

func (c *Crawler) WriteTemporaryFile() {
	if len(c.ProblemsToBeReprocessed) > 0 {
		tempFile, err := os.Create(TEMP_FILE_PATH)
		if err != nil {
			log.Fatal("Failed to create temporary file:", err)
		}
		defer tempFile.Close()

		for _, entry := range c.ProblemsToBeReprocessed {
			if _, err := tempFile.WriteString(entry + "\n"); err != nil {
				log.Fatal("Failed to write to temporary file:", err)
			}
			log.Println("Record temporary code: " + entry)
		}
	}
}

func (c *Crawler) Scraping() {
	pageNum := START_PAGE
	for {
		submissionList, err := c.Lc.GetSubmissionList(pageNum)
		if err != nil {
			log.Println("Failed to get submission list:", err)
			break
		}
		expired := c.ProcessSubmissions(submissionList.Submissions)
		if !submissionList.HasNext || expired {
			log.Println("No more submissions!")
			break
		}
		pageNum += LIMIT
		time.Sleep(PAGE_TIME * time.Second)
	}
	c.ProcessTemporaryProblems()
	c.WriteTemporaryFile()
}

func (c *Crawler) Execute() {
	log.Println("Login")
	if err := c.Lc.Login(); err != nil {
		log.Fatal("Login failed:", err)
	}
	log.Println("Start scraping")
	c.Scraping()
	log.Println("End scraping")
	if err := utils.GitPush(c.PushDir); err != nil {
		log.Fatal("Git push failed:", err)
	}
}
