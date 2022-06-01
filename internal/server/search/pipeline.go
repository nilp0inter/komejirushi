package search

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/junegunn/fzf/src/algo"
	"github.com/junegunn/fzf/src/util"
	"github.com/nilp0inter/komejirushi/internal/server/commands"
	"github.com/nilp0inter/komejirushi/internal/server/config"
)

func merge(cs ...<-chan commands.TaggedSearchResult) <-chan commands.TaggedSearchResult {
	var wg sync.WaitGroup
	out := make(chan commands.TaggedSearchResult)

	// Start an output goroutine for each input channel in cs.  output
	// copies values from c to out until c is closed, then calls wg.Done.
	output := func(c <-chan commands.TaggedSearchResult) {
		for n := range c {
			out <- n
		}
		wg.Done()
	}
	wg.Add(len(cs))
	for _, c := range cs {
		go output(c)
	}

	// Start a goroutine to close out once all the output goroutines are
	// done.  This must start after the wg.Add call.
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

func score(re, s string) int {
	chars := util.ToChars([]byte(s))
	res, _ := algo.FuzzyMatchV2(false, false, false, &chars, []rune(re), true, nil)
	if res.Start < 0 {
		return -1
	}
	return res.Score
}

func MakeSearch(c config.Config, term string, out chan<- commands.SearchResponse) {
	var results []<-chan commands.TaggedSearchResult
	now := time.Now()

	for ds, db := range c.Docsets {
		rs := func(ds string, db *sql.DB) <-chan commands.TaggedSearchResult {
			entries := make(chan commands.TaggedSearchResult)
			go func() {
				defer close(entries)
				rows, err := db.Query("select name from searchIndex where subsetchrs(?, name)", term)
				if err != nil {
					log.Fatal(err)
				}
				defer rows.Close()

				for rows.Next() {
					var name string
					err = rows.Scan(&name)
					if err != nil {
						log.Fatal(err)
					}
					s := score(term, name)
					if s > -1 {
						entries <- commands.TaggedSearchResult{
							Docset: ds,
							Result: commands.SearchResult{Name: name, Score: s},
						}
					}
				}
				err = rows.Err()
				if err != nil {
					log.Fatal(err)
				}
			}()
			return entries
		}(ds, db)
		results = append(results, rs)
	}
	t := 0

	chunk := make(map[string][]commands.SearchResult)
	ticker := time.NewTicker(50 * time.Millisecond)
	merged := merge(results...)
R:
	for {
		select {
		case r, ok := <-merged:
			if !ok {
				ticker.Stop()
				break R
			}
			t++
			chunk[r.Docset] = append(chunk[r.Docset], r.Result)
		case <-ticker.C:
			fmt.Println("partial:", t)
			t = 0
			if len(chunk) != 0 {
				out <- commands.SearchResponse{Results: chunk}
				chunk = make(map[string][]commands.SearchResult)
			}
		}
	}
	if len(chunk) != 0 {
		out <- commands.SearchResponse{Results: chunk}
		chunk = make(map[string][]commands.SearchResult)
	}
	log.Println("search:", term, "time elapse:", time.Since(now), "total:", t)
}
