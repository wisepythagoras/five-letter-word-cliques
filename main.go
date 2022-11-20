package main

import (
	"flag"
	"fmt"
	"math/bits"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

var mainChan chan []uint32
var reqChan chan int
var jobChans [](chan uint32)
var wg sync.WaitGroup

func getWordMask(word string) uint32 {
	mask := uint32(0)

	for _, char := range word {
		mask |= 1 << (char - 'a')
	}

	return mask
}

func hasDuplicateLetter(word string) (bool, uint32) {
	mask := getWordMask(word)

	return bits.OnesCount32(mask) != len(word), mask
}

func hasUniqueLetters(word1 uint32, word2 uint32) bool {
	return word1&word2 == 0
}

func findFiveWords(wordMasks []uint32, word uint32, count uint8, tmp []uint32, res *[][]uint32, findAll bool) uint8 {
	if count == 4 {
		*res = append(*res, tmp)

		// Found 5 words.
		if findAll {
			return count - 1
		} else {
			return count + 1
		}
	}

	for _, mask := range wordMasks {
		if hasUniqueLetters(mask, word) {
			newCount := findFiveWords(wordMasks, word|mask, count+1, append(tmp, mask), res, findAll)

			if newCount > count {
				return count
			}
		}
	}

	return count - 1
}

func childRoutine(words []uint32, id int, jobChan <-chan uint32, findAll bool, verbose bool) {
	defer wg.Done()

	for word := range jobChan {
		if word == 0 {
			continue
		}

		results := make([][]uint32, 0)
		findFiveWords(words, word, 0, []uint32{word}, &results, findAll)

		// Request a new job.
		reqChan <- id

		for _, result := range results {
			mainChan <- result
		}
	}

	vPrintln(verbose, id, " Done")
}

func vPrintln(verbose bool, args ...any) {
	if !verbose {
		return
	}

	fmt.Println(args...)
}

func main() {
	verbosePtr := flag.Bool("verbose", false, "Print out all messages")
	wordFilePtr := flag.String("word-file", "", "The path to the word file")
	outputListPtr := flag.Bool("output-list", false, "Outputs the list of words without repeating letters")
	findAllPtr := flag.Bool("find-all", false, "Find all combinations (not just the first)")

	flag.Parse()

	if wordFilePtr == nil || *wordFilePtr == "" {
		fmt.Println("A word file is required: -word-file path/to/word/file")
		os.Exit(1)
	}

	byteContents, err := os.ReadFile(*wordFilePtr)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	verbose := *verbosePtr || *outputListPtr
	cores := runtime.NumCPU()
	mainChan = make(chan []uint32)
	reqChan = make(chan int)
	jobChans = make([](chan uint32), 0)
	contents := string(byteContents)
	words := strings.Split(contents, "\n")
	wordMasks := make(map[uint32][]string)
	masks := make([]uint32, 0)
	sort.Strings(words)

	start := time.Now()

	for _, word := range words {
		if has, bits := hasDuplicateLetter(word); !has && bits != 0 {
			_, wordIsInMap := wordMasks[bits]

			vPrintln(verbose, word, bits)

			wordMasks[bits] = append(wordMasks[bits], word)

			if !wordIsInMap {
				masks = append(masks, bits)
			}
		}
	}

	if *outputListPtr {
		os.Exit(0)
	}

	for i := 0; i < cores; i++ {
		jobChans = append(jobChans, make(chan uint32, 1024))
		jobChans[i] <- masks[i]
	}

	go func() {
		// We've already sent the first jobs to the workers.
		i := cores
		openJobs := cores

		// Listen for requests and fulfill them by sending the next word in the list.
		for reqId := range reqChan {
			if i >= len(masks) {
				vPrintln(verbose, "Close jobChan ", reqId)

				close(jobChans[reqId])
				openJobs--

				if openJobs == 0 {
					close(reqChan)
					break
				}

				continue
			}

			if verbose {
				fmt.Printf("%d <- %s (%d) @%d\n", reqId, wordMasks[masks[i]], masks[i], i)
			}

			jobChans[reqId] <- masks[i]
			i++
		}
	}()

	vPrintln(verbose, "Check", len(masks), "words")

	for i := 0; i < cores; i++ {
		wg.Add(1)
		go childRoutine(masks, i, jobChans[i], *findAllPtr, verbose)
	}

	go func() {
		wg.Wait()
		vPrintln(verbose, "Shut down")
		close(mainChan)

		duration := time.Since(start)
		ms := duration.Milliseconds()

		fmt.Print("Execution time: ")

		if ms > 0 {
			fmt.Printf("%dms\n", ms)
		} else {
			fmt.Printf("0.%dms\n", duration.Microseconds())
		}
	}()

	for foundMasks := range mainChan {
		var strBuilder strings.Builder

		for i, mask := range foundMasks {
			if i > 0 {
				strBuilder.WriteString(",")
			}

			strBuilder.WriteString(strings.Join(wordMasks[mask], "|"))
		}

		fmt.Println(strBuilder.String())
	}
}
