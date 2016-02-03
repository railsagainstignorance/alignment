package main

import (
    "bufio"
    "os"
    "strings"
    "fmt"
    "regexp"
    "sort"
)

const (
	SyllableFilename = "./cmudict-0.7b"
)

type Word struct {
    Name            string
    FragmentsString string
    Fragments       []string
    NumSyllables    int
    FinalSyllable   string
    EmphasisPoints  []string
}

var (
	syllableRegexp      = regexp.MustCompile(`^[A-Z]+(\d+)$`)
	finalSyllableRegexp = regexp.MustCompile(`([A-Z]+\d+(?:[^\d]*))$`)
)

func readSyllables(filename string) (*map[string]*Word, int, int) {

	words := map[string]*Word{}
	countFragments      := 0
	countSyllables      := 0
    // syllableRegexp      := regexp.MustCompile(`^[A-Z]+\d+$`)
    // FinalSyllableRegexp := regexp.MustCompile(`([A-Z]+\d+(?:[^\d]*))$`)

    // Open the file.
    f, _ := os.Open(filename)
    // Create a new Scanner for the file.
    scanner := bufio.NewScanner(f)
    // Loop over all lines in the file
    for scanner.Scan() {
		line := scanner.Text()
		if ! strings.HasPrefix(line, ";;;") {
			nameAndRemainder := strings.Split(line, "  ")
			name             := nameAndRemainder[0]
			remainder        := nameAndRemainder[1]
			fragments        := strings.Split(remainder, " ")
			emphasisPoints   := []string{}

			numSyllables := 0
			for _,f := range fragments {
				matches := syllableRegexp.FindStringSubmatch(f)
				if matches != nil {
					numSyllables = numSyllables + 1
					emphasisPoints = append(emphasisPoints, matches[1])
	    		}
	    	}

	    	if numSyllables == 0 {
	    		fmt.Println("WARNING: no syllables found for name=", name) 
	    	}

	    	matches := finalSyllableRegexp.FindStringSubmatch(remainder)
	    	finalSyllable := ""
	    	if matches != nil {
	    		finalSyllable = matches[1]
	    	} else {
	    		fmt.Println("WARNING: no final syllable found for name=", name) 
	    	}

	    	countSyllables = countSyllables + numSyllables
			countFragments = countFragments + len(fragments)
			words[name] = &Word{
				Name:            name,
				FragmentsString: remainder,
				Fragments:       fragments,
				NumSyllables:    numSyllables,
				FinalSyllable:   finalSyllable,
				EmphasisPoints:  emphasisPoints,
			}
		}
    }

    // fmt.Println("num fragments = ", countFragments, ", num syllables = ", countSyllables) 

    return &words, countFragments, countSyllables
}

func processFinalSyllables(words *map[string]*Word) (*map[string][]*Word) {
	finalSyllables := map[string][]*Word{}

	for _,word := range *words {
		fs := word.FinalSyllable
		var rhymingWords []*Word

		rhymingWords, ok := finalSyllables[fs]
		if ! ok {
			rhymingWords = []*Word{}
		}

		finalSyllables[fs] = append( rhymingWords, word )
	}

	return &finalSyllables
}

type Stats struct {
	NumWords                int
	NumUniqueFinalSyllables int
	NumFragments            int
	NumSyllables            int 
}

type Syllabi struct {
	Stats          Stats
    SourceFilename string
    FindRhymes     func(string) []string
    CountSyllables func(string) int
    EmphasisPoints func(string) []string
    FinalSyllable  func(string) string
}

func ConstructSyllabi(sourceFilename string) (*Syllabi){
	words, numFragments, numSyllables := readSyllables(SyllableFilename)
	finalSyllables := processFinalSyllables(words)

	stats := Stats{
		NumWords:                len(*words),
		NumUniqueFinalSyllables: len( *finalSyllables),
		NumFragments:            numFragments,
		NumSyllables:            numSyllables,
	}

	findRhymes := func(s string) []string {
		upperS          := strings.ToUpper(s)
		matchingStrings := []string{}

	    // fmt.Println("ConstructSyllabi: upperS=", upperS ) 

		if matchingWord, ok := (*words)[upperS]; ok {
		     // fmt.Println("ConstructSyllabi: matchingWord=", matchingWord ) 
		     // fmt.Println("ConstructSyllabi: matchingWord.FragmentsString=", matchingWord.FragmentsString ) 
		     // fmt.Println("ConstructSyllabi: matchingWord.FinalSyllable=", matchingWord.FinalSyllable ) 
			finalSyllable := matchingWord.FinalSyllable
		 	if rhymingWords, ok := (*finalSyllables)[finalSyllable]; ok {
		 		for _,w := range rhymingWords {
		 			matchingStrings = append(matchingStrings, (*w).Name)
				}
			}
		}

		return matchingStrings
	}

	countSyllables := func(s string) int {
		upperS := strings.ToUpper(s)
		count  := 0

		if w, ok := (*words)[upperS]; ok {
			count = (*w).NumSyllables
		}

		return count
	}

	emphasisPoints := func(s string) []string {
		upperS := strings.ToUpper(s)
		ep := []string{}
		if w,ok := (*words)[upperS]; ok {
			ep = (*w).EmphasisPoints
		}
		return ep
	}

	finalSyllableFunc := func(s string) string {
		upperS := strings.ToUpper(s)
		fs := ""
		if w,ok := (*words)[upperS]; ok {
			fs = (*w).FinalSyllable
		}
		return fs
	}

	syllabi := Syllabi{
		Stats:          stats,
		SourceFilename: SyllableFilename,
		FindRhymes:     findRhymes,
		CountSyllables: countSyllables,
		EmphasisPoints: emphasisPoints,
		FinalSyllable:  finalSyllableFunc,
	}

	return &syllabi
}

func main() {
	syllabi := ConstructSyllabi(SyllableFilename)

    fmt.Println("main: num words =", (*syllabi).Stats.NumWords ) 
	fmt.Println("main: num unique final syllables =", (*syllabi).Stats.NumUniqueFinalSyllables)
	fmt.Println("main: num fragments =", (*syllabi).Stats.NumFragments)
	fmt.Println("main: num syllables =", (*syllabi).Stats.NumSyllables)

	s := "hyperactivity"
	rhymesWith := (*syllabi).FindRhymes(s)
	sort.Strings(rhymesWith)
	fmt.Println("main:", s, "rhymes with", len(rhymesWith), ": first=", rhymesWith[0], ", last=", rhymesWith[len(rhymesWith)-1])

	numSyllables := (*syllabi).CountSyllables(s)
	fmt.Println("main:", s, "has", numSyllables, "syllables")

	ep := (*syllabi).EmphasisPoints(s)
	fmt.Println("main:", s, "emphasisPoints=", strings.Join(ep, ","))

	fs := (*syllabi).FinalSyllable(s)
	fmt.Println("main:", s, "finalSyllable=", fs)
}
