package main

import (
	"fmt"
	"github.com/netscrn/homm3utils/defparse"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

func main() {
	if len(os.Args) > 3 {
		panic("invalid arguments")
	}

	dirContent, err := os.ReadDir(os.Args[1])
	if err != nil {
		panic(err)
	}

	concurrencyLevel := 1
	var wg sync.WaitGroup
	wg.Add(concurrencyLevel)

	batchSize := len(dirContent) / concurrencyLevel
	remainingFiles := len(dirContent) % concurrencyLevel
	for l := 1; l <= concurrencyLevel; l++ {
		start := batchSize * (l-1)
		end   := batchSize * l
		if (l == concurrencyLevel) && (remainingFiles != 0) {
			end += remainingFiles
		}

		go func() {
			defer wg.Done()

			for _, entry := range dirContent[start:end] {
				if strings.HasSuffix(entry.Name(), ".def") {
					err := defparse.ExtractDef(filepath.Join(os.Args[1], entry.Name()), os.Args[2])
					if err != nil {
						fmt.Printf("Can't extract %s: %v\n", entry.Name(), err)
						continue
					}
				}
			}
		}()
	}
	wg.Wait()
}