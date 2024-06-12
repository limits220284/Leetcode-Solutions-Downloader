package main

import (
	"flag"
	"log"
)

type Args struct {
	Cookie    string
	Output    string
	Day       int
	Overwrite bool
}

func main() {
	cookie := flag.String("c", "", "Your cookie for login")
	output := flag.String("o", "", "Output path")
	day := flag.Int("d", 0, "Fetching codes in 'day'")
	overwrite := flag.Bool("O", false, "Flag to enable overwrite")

	flag.Parse()

	args := Args{
		Cookie:    *cookie,
		Output:    *output,
		Day:       *day,
		Overwrite: *overwrite,
	}

	// Execute the crawler once
	crawler := NewCrawler(args)
	crawler.Execute()

	// Uncomment to run the crawler every 24 hours
	// c := cron.New()
	// _, err := c.AddFunc("@every 24h", func() {
	// 	crawler := NewCrawler(args)
	// 	crawler.Execute()
	// })
	// if err != nil {
	// 	log.Fatalf("Error scheduling the job: %v", err)
	// }
	// c.Start()
	// defer c.Stop()

	// Wait indefinitely
	// select {}
}

// NewCrawler is a placeholder function. Replace it with your actual Crawler constructor.
func NewCrawler(args Args) *Crawler {
	// Implement this function based on your actual Crawler constructor in Go.
	return &Crawler{
		Args: args,
	}
}

// Crawler is a placeholder struct. Replace it with your actual Crawler struct.
type Crawler struct {
	Args Args
}

// Execute is a placeholder method. Replace it with your actual Execute method.
func (c *Crawler) Execute() {
	// Implement this method based on your actual Execute method in Go.
	log.Println("Executing the crawler with args:", c.Args)
}
