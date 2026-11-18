// Command genro builds the Romanian wordlist (internal/wordlist/ro_diceware.txt).
//
// It counts how often words occur in a Romanian Wikipedia dump and keeps the
// most frequent ones that are base forms in the rospell dictionary, are 3 to 7
// letters without diacritics, and are not on the blocklist. The words are
// sorted and numbered with the dice rolls 11111 to 66666. See NOTICE for the
// sources and their licences.
//
//	go run ./scripts/genro -wiki rowiki.xml.bz2 -dic index.dic -out internal/wordlist/ro_diceware.txt
package main

import (
	"bufio"
	"compress/bzip2"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strings"
	"unicode"
)

const listSize = 7776

var (
	candidate = regexp.MustCompile(`^[a-z]{3,7}$`)
	// Wiki markup that would otherwise give words like "ref" or "thumb" a
	// high count.
	markup = []*regexp.Regexp{
		regexp.MustCompile(`(?s)<!--.*?-->`),
		regexp.MustCompile(`(?s)<ref[^>]*>.*?</ref>`),
		regexp.MustCompile(`<[^>]*>`),
		regexp.MustCompile(`https?://\S+`),
		regexp.MustCompile(`\[\[(?:Fișier|Fişier|File|Imagine|Image|Categorie|Category):[^\]]*\]\]`),
		regexp.MustCompile(`\|\s*[A-Za-z_ -]+\s*=`),
	}
	template = regexp.MustCompile(`\{\{[^{}]*\}\}`)
	// Endings of plural and definite-article forms, which the dictionary also
	// lists as headwords.
	inflected = []string{"ile", "ilor", "elor", "ului", "ilei", "urile", "lor"}
	roman     = regexp.MustCompile(`^[ivxlcdm]+$`)
)

// minLowerShare is the least share of a word's occurrences that must start
// with a lowercase letter. Names that the dictionary also lists in lowercase
// (brands, places) are capitalized almost every time.
const minLowerShare = 0.9

func main() {
	wiki := flag.String("wiki", "", "Romanian Wikipedia pages-articles .xml.bz2 (may be truncated)")
	dic := flag.String("dic", "", "rospell index.dic")
	blockPath := flag.String("block", "scripts/genro/blocklist.txt", "blocklist")
	out := flag.String("out", "", "output wordlist")
	countsPath := flag.String("counts", "scripts/genro/counts.tsv", "word counts: read if it exists, else made from -wiki and written")
	share := flag.Float64("share", minLowerShare, "least share of occurrences that start with a lowercase letter")
	minCount := flag.Int("min", 20, "least number of occurrences")
	flag.Parse()
	if *dic == "" || *out == "" {
		flag.Usage()
		os.Exit(2)
	}
	if err := run(*wiki, *dic, *blockPath, *out, *countsPath, *minCount, *share); err != nil {
		fmt.Fprintln(os.Stderr, "genro:", err)
		os.Exit(1)
	}
}

func run(wikiPath, dicPath, blockPath, outPath, countsPath string, minCount int, share float64) error {
	stems, err := readStems(dicPath)
	if err != nil {
		return err
	}
	block, err := readBlocklist(blockPath)
	if err != nil {
		return err
	}
	counts, err := loadCounts(countsPath)
	if errors.Is(err, os.ErrNotExist) {
		if wikiPath == "" {
			return fmt.Errorf("%s does not exist; give -wiki to make it", countsPath)
		}
		if counts, err = countWords(wikiPath, stems); err == nil {
			err = saveCounts(countsPath, counts)
		}
	}
	if err != nil {
		return err
	}

	type wc struct {
		word  string
		count int
	}
	var cands []wc
	for w, n := range counts {
		if n.lower >= minCount && n.lowerShare() >= share && stems[w] && !block[w] && !isInflected(w) && !isArticled(w, stems) && !roman.MatchString(w) {
			cands = append(cands, wc{w, n.lower})
		}
	}
	// Frequent words first; ties by spelling so the output is stable.
	sort.Slice(cands, func(i, j int) bool {
		if cands[i].count != cands[j].count {
			return cands[i].count > cands[j].count
		}
		return cands[i].word < cands[j].word
	})
	fmt.Fprintf(os.Stderr, "%d candidates; need %d (least frequent kept: %d)\n", len(cands), listSize, cands[min(listSize, len(cands))-1].count)
	if len(cands) < listSize {
		return fmt.Errorf("only %d candidates, need %d: lower -min or add sources", len(cands), listSize)
	}
	words := make([]string, listSize)
	for i := range words {
		words[i] = cands[i].word
	}
	sort.Strings(words)

	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	w := bufio.NewWriter(f)
	for i, word := range words {
		_, _ = fmt.Fprintf(w, "%s\t%s\n", diceRoll(i), word)
	}
	if err := w.Flush(); err != nil {
		return err
	}
	return f.Close()
}

// diceRoll renders index 0..7775 as five digits 1-6, from 11111 to 66666.
func diceRoll(i int) string {
	var d [5]byte
	for p := 4; p >= 0; p-- {
		d[p] = byte('1' + i%6)
		i /= 6
	}
	return string(d[:])
}

// isArticled reports whether w is another stem plus the definite article or
// a plural ending ("dreptul" from "drept", "marea" from "mare").
func isArticled(w string, stems map[string]bool) bool {
	for _, suffix := range []string{"ul", "le", "ii", "ei"} {
		if strings.HasSuffix(w, suffix) && stems[strings.TrimSuffix(w, suffix)] {
			return true
		}
	}
	return strings.HasSuffix(w, "ea") && stems[strings.TrimSuffix(w, "a")]
}

func isInflected(w string) bool {
	for _, suffix := range inflected {
		if strings.HasSuffix(w, suffix) {
			return true
		}
	}
	return false
}

// readStems returns the lowercase headwords of a hunspell .dic file that can
// be wordlist entries.
func readStems(path string) (map[string]bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	stems := map[string]bool{}
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<20), 1<<20)
	first := true
	for sc.Scan() {
		if first { // the entry count
			first = false
			continue
		}
		w, _, _ := strings.Cut(sc.Text(), "/")
		if candidate.MatchString(w) {
			stems[w] = true
		}
	}
	return stems, sc.Err()
}

func readBlocklist(path string) (map[string]bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block := map[string]bool{}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			block[line] = true
		}
	}
	return block, nil
}

// countWords counts the occurrences of every stem in the dump. A truncated
// dump is fine: the part that is there is counted.
type tally struct{ lower, capital int }

func (t tally) lowerShare() float64 {
	return float64(t.lower) / float64(t.lower+t.capital)
}

func countWords(path string, stems map[string]bool) (map[string]*tally, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	r := bufio.NewReaderSize(bzip2.NewReader(f), 1<<20)

	counts := map[string]*tally{}
	var total int
	for {
		// The dump is one line per text line; a page's text spans many.
		line, err := r.ReadString('\n')
		for _, re := range markup {
			line = re.ReplaceAllString(line, " ")
		}
		for prev := ""; prev != line; {
			prev = line
			line = template.ReplaceAllString(line, " ")
		}
		for _, raw := range strings.FieldsFunc(line, func(c rune) bool { return !unicode.IsLetter(c) }) {
			total++
			tok := strings.ToLower(raw)
			if !stems[tok] {
				continue
			}
			t := counts[tok]
			if t == nil {
				t = &tally{}
				counts[tok] = t
			}
			if unicode.IsLower([]rune(raw)[0]) {
				t.lower++
			} else {
				t.capital++
			}
		}
		if err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
				break
			}
			return nil, err
		}
	}
	fmt.Fprintf(os.Stderr, "%d tokens, %d distinct stems seen\n", total, len(counts))
	return counts, nil
}

// saveCounts and loadCounts keep the counts in a small text file, so the list
// can be rebuilt (and the selection rules tuned) without the 800 MB dump.
func saveCounts(path string, counts map[string]*tally) error {
	words := make([]string, 0, len(counts))
	for w := range counts {
		words = append(words, w)
	}
	sort.Strings(words)
	var b strings.Builder
	b.WriteString("# word, occurrences starting lowercase, occurrences starting uppercase\n")
	for _, w := range words {
		fmt.Fprintf(&b, "%s\t%d\t%d\n", w, counts[w].lower, counts[w].capital)
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

func loadCounts(path string) (map[string]*tally, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	counts := map[string]*tally{}
	for _, line := range strings.Split(string(data), "\n") {
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		var w string
		var t tally
		if _, err := fmt.Sscanf(line, "%s\t%d\t%d", &w, &t.lower, &t.capital); err != nil {
			return nil, fmt.Errorf("%s: bad line %q: %w", path, line, err)
		}
		counts[w] = &t
	}
	return counts, nil
}
