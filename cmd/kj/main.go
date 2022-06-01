package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/junegunn/fzf/src/algo"
	"github.com/junegunn/fzf/src/util"
	"github.com/nilp0inter/komejirushi/internal/server"
)

func searchAgent(path string, qs chan string, rs chan string) {
	db, err := sql.Open("sqlite3_extended", path)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	for query := range qs {
		fmt.Println(path, query)
		func() {
			rows, err := db.Query("select name from searchIndex where subsetchrs(?, name)", query)
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
				rs <- name
				// sc, _ := score(strings.ToLower(query), name)
				// if sc >= best {
				// 	bname = name
				// 	best = sc
				// }
			}
			err = rows.Err()
			if err != nil {
				log.Fatal(err)
			}
		}()
	}
}

func removeDuplicateRunes(runeSlice []rune) []rune {
	keys := make(map[rune]bool)
	list := []rune{}

	// If the key(values of the slice) is not equal
	// to the already present value in new slice (list)
	// then we append it. else we jump on another element.
	for _, entry := range runeSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}
func score(re, s string) (int, error) {
	// fun(caseSensitive, normalize, forward, &chars, []rune(pattern), true, nil)
	// func FuzzyMatchV2(caseSensitive bool, normalize bool, forward bool, input *util.Chars, pattern []rune, withPos bool, slab *util.Slab) (Result, *[]int) {
	chars := util.ToChars([]byte(s))
	res, _ := algo.FuzzyMatchV2(false, false, false, &chars, []rune(re), true, nil)
	if res.Start < 0 {
		return 0, nil
	}
	return res.Score, nil
}

func main() {
	server.RunServer(os.Args[1:]...)
}
