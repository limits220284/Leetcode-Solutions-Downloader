package main

import (
	"flag"
	"lsc/pkg/crawler"
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

	args := crawler.Args{
		Cookie:    *cookie,
		Output:    *output,
		Day:       *day,
		Overwrite: *overwrite,
	}

	// Execute the crawler once
	crawler := crawler.NewCrawler(args)
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
