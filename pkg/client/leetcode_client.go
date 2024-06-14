package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// 定义与 JSON 数据结构对应的 Go 结构体
type Submission struct {
	ID            int    `json:"id"`
	Lang          string `json:"lang"`
	Time          string `json:"time"`
	StatusDisplay string `json:"status_display"`
	Runtime       string `json:"runtime"`
	URL           string `json:"url"`
	IsPending     string `json:"is_pending"`
	Title         string `json:"title"`
	Timestamp     int64  `json:"timestamp"`
	Memory        string `json:"memory"`
}

type SubmissionsDump struct {
	Submissions []Submission `json:"submissions_dump"`
	HasNext     bool         `json:"has_next"`
	LastKey     string       `json:"last_key"`
}

type SubmissionDetailResponse struct {
	Code          string `json:"data.submissionDetail.code"`
	QuestionID    string `json:"data.submissionDetail.question.questionFrontendId"`
	QuestionTitle string `json:"data.submissionDetail.question.translatedTitle"`
}

type LeetcodeClient struct {
	Client    *http.Client
	Cookie    string
	SleepTime time.Duration
	Endpoint  string
	Headers   map[string]string
}

func NewLeetcodeClient(cookie string) *LeetcodeClient {
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
			log.Println("Login successfully!")
			return nil
		}
		log.Println("Login failed, Wait till next round!")
		if tryCnt != ATTEMPT-1 {
			time.Sleep(lc.SleepTime)
		}
	}

	log.Println("LoginError: Login failed, ensure your login credential is correct!")
	return errors.New("LoginError: Login failed, ensure your login credential is correct!")
}

func (lc *LeetcodeClient) DownloadCode(submission Submission) (SubmissionDetailResponse, error) {
	queryFile, err := os.ReadFile("query/query_download_submission")
	if err != nil {
		return SubmissionDetailResponse{}, fmt.Errorf("failed to read query file: %w", err)
	}

	queryString := string(queryFile)
	data := map[string]interface{}{
		"query":         queryString,
		"operationName": "mySubmissionDetail",
		"variables":     map[string]interface{}{"id": submission.ID},
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return SubmissionDetailResponse{}, fmt.Errorf("failed to marshal JSON: %w", err)
	}

	req, err := http.NewRequest("POST", lc.Endpoint+"graphql/", bytes.NewBuffer(jsonData))
	if err != nil {
		return SubmissionDetailResponse{}, fmt.Errorf("failed to create request: %w", err)
	}
	for key, value := range lc.Headers {
		req.Header.Set(key, value)
	}
	resp, err := lc.Client.Do(req)
	if err != nil {
		return SubmissionDetailResponse{}, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return SubmissionDetailResponse{}, fmt.Errorf("failed to read response body: %w", err)
	}

	var response SubmissionDetailResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return SubmissionDetailResponse{}, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// submissionDetails, ok := response["data"].(map[string]interface{})["submissionDetail"]
	// submissionDetails, ok := response.Co
	// if !ok {
	// 	return nil, errors.New("failed to retrieve submission details")
	// }

	return response, nil
}

func (lc *LeetcodeClient) GetSubmissionList(pageNum int) (SubmissionsDump, error) {
	log.Printf("Now scraping submissions list for page: %d\n", pageNum)
	submissionsURL := fmt.Sprintf("https://leetcode.cn/api/submissions/?offset=%d&limit=40", pageNum)
	req, err := http.NewRequest("GET", submissionsURL, nil)
	if err != nil {
		return SubmissionsDump{}, fmt.Errorf("failed to create request: %w", err)
	}
	for key, value := range lc.Headers {
		req.Header.Set(key, value)
	}
	resp, err := lc.Client.Do(req)
	if err != nil {
		return SubmissionsDump{}, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return SubmissionsDump{}, fmt.Errorf("failed to read response body: %w", err)
	}

	var submissionsList SubmissionsDump
	if err := json.Unmarshal(body, &submissionsList); err != nil {
		return SubmissionsDump{}, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	log.Println(submissionsList)
	return submissionsList, nil
}
