package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

type LeetcodeClient struct {
	Client    *http.Client
	Cookie    string
	SleepTime time.Duration
	Endpoint  string
	Logger    *log.Logger
	Headers   map[string]string
}

func NewLeetcodeClient(cookie string, logger *log.Logger) *LeetcodeClient {
	client := &http.Client{}
	headers := map[string]string{
		"Connection":   "keep-alive",
		"Content-Type": "application/json",
		"User-Agent":   "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_9_2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/33.0.1750.152 Safari/537.36",
	}

	return &LeetcodeClient{
		Client:    client,
		Cookie:    cookie,
		SleepTime: 5 * time.Second,
		Endpoint:  "https://leetcode.cn/",
		Logger:    logger,
		Headers:   headers,
	}
}

func (lc *LeetcodeClient) Login() error {
	const ATTEMPT = 3
	loginURL := lc.Endpoint + "accounts/login/"
	loginHeader := lc.Headers
	loginHeader["Referer"] = loginURL

	for tryCnt := 0; tryCnt < ATTEMPT; tryCnt++ {
		req, err := http.NewRequest("GET", loginURL, nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("Cookie", "LEETCODE_SESSION="+lc.Cookie)
		resp, err := lc.Client.Do(req)
		if err != nil {
			return fmt.Errorf("failed to execute request: %w", err)
		}
		resp.Body.Close()

		req, err = http.NewRequest("POST", loginURL, nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("Cookie", "LEETCODE_SESSION="+lc.Cookie)
		for key, value := range loginHeader {
			req.Header.Set(key, value)
		}
		resp, err = lc.Client.Do(req)
		if err != nil {
			return fmt.Errorf("failed to execute request: %w", err)
		}
		if resp.StatusCode == http.StatusOK && resp.Request.URL.String() == lc.Endpoint {
			lc.Logger.Println("Login successfully!")
			return nil
		}
		lc.Logger.Println("Login failed, Wait till next round!")
		if tryCnt != ATTEMPT-1 {
			time.Sleep(lc.SleepTime)
		}
	}

	lc.Logger.Println("LoginError: Login failed, ensure your login credential is correct!")
	return errors.New("LoginError: Login failed, ensure your login credential is correct!")
}

func (lc *LeetcodeClient) DownloadCode(submission map[string]interface{}) (map[string]interface{}, error) {
	queryFile, err := os.ReadFile("query/query_download_submission")
	if err != nil {
		return nil, fmt.Errorf("failed to read query file: %w", err)
	}

	queryString := string(queryFile)
	data := map[string]interface{}{
		"query":         queryString,
		"operationName": "mySubmissionDetail",
		"variables":     map[string]interface{}{"id": submission["id"]},
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}

	req, err := http.NewRequest("POST", lc.Endpoint+"graphql/", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	for key, value := range lc.Headers {
		req.Header.Set(key, value)
	}
	resp, err := lc.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	submissionDetails, ok := response["data"].(map[string]interface{})["submissionDetail"]
	if !ok {
		return nil, errors.New("failed to retrieve submission details")
	}

	return submissionDetails.(map[string]interface{}), nil
}

func (lc *LeetcodeClient) GetSubmissionList(pageNum int) (map[string]interface{}, error) {
	lc.Logger.Printf("Now scraping submissions list for page: %d\n", pageNum)
	submissionsURL := fmt.Sprintf("https://leetcode.cn/api/submissions/?offset=%d&limit=40", pageNum)
	req, err := http.NewRequest("GET", submissionsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	for key, value := range lc.Headers {
		req.Header.Set(key, value)
	}
	resp, err := lc.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var submissionsList map[string]interface{}
	if err := json.Unmarshal(body, &submissionsList); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return submissionsList, nil
}
