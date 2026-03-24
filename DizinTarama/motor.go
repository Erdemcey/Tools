package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

var client = &http.Client{
	Timeout: 6 * time.Second,
}

func readWordlist(path string) []string {
	file, err := os.Open(path)
	if err != nil {
		fmt.Println("wordlist açılamadı:", path)
		return nil
	}
	defer file.Close()

	var list []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		list = append(list, line)
	}

	return list
}

func worker(id int, jobs <-chan string, wg *sync.WaitGroup) {
	defer wg.Done()

	for target := range jobs {

		req, err := http.NewRequest("GET", target, nil)
		if err != nil {
			continue
		}

		req.Header.Set("User-Agent", "Mozilla/5.0")

		resp, err := client.Do(req)
		if err != nil {
			continue
		}

		// basit filtre
		if resp.StatusCode != 404 && resp.StatusCode != 400 {
			fmt.Printf("[%d] %s\n", resp.StatusCode, target)
		}

		resp.Body.Close()
	}
}

func main() {
	url := flag.String("u", "", "target (örn: http://site.com/[[dir]])")
	subFile := flag.String("s", "", "subdomain list")
	dirFile := flag.String("d", "", "dir list")
	thread := flag.Int("t", 30, "thread")
	flag.Parse()

	if *url == "" {
		fmt.Println("kullanım: -u http://site.com/[[dir]]")
		return
	}

	subs := readWordlist(*subFile)
	dirs := readWordlist(*dirFile)

	if len(subs) == 0 {
		subs = []string{""}
	}
	if len(dirs) == 0 {
		dirs = []string{""}
	}

	jobs := make(chan string, *thread)
	var wg sync.WaitGroup

	// workerlar
	for i := 0; i < *thread; i++ {
		wg.Add(1)
		go worker(i, jobs, &wg)
	}

	total := 0

	for _, s := range subs {
		for _, d := range dirs {

			target := *url
			target = strings.ReplaceAll(target, "[[sub]]", s)
			target = strings.ReplaceAll(target, "[[dir]]", d)

			jobs <- target
			total++
		}
	}

	close(jobs)

	fmt.Println("yüklenen iş:", total)

	wg.Wait()

	fmt.Println("bitti.")
}
