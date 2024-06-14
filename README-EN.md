# LeetCode Solution Downloader
A LeetCode solution downloader. Welcome to raise issues and submit pull requests.

# Usage
1. Configure config.json
```json
{
    "output_dir": "./output", // Output directory for downloaded files
    "cookie": "***", // Your LeetCode login cookie
    "day": 100 // Download solutions for the last 100 days, you can customize this
}
```
2. Run
```shell
go run main.go
```

# TODO
- [ ] Utilize Golang's concurrency to improve download speed
- [ ] Containerize operations, periodically crawl solutions, and push them to the repository
- [ ] Add timeout retry mechanism. Retry if crawling fails.