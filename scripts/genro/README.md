# genro: the Romanian wordlist

`go run ./scripts/genro` writes `internal/wordlist/ro_diceware.txt`: 7,776
common Romanian words, 3 to 7 letters, no diacritics, numbered with the dice
rolls 11111 to 66666.

A word is kept when it

* is a base form (headword) in the rospell dictionary, so names, typos and
  most inflected forms are out;
* occurs at least 10 times in the Romanian Wikipedia, and at least half of
  those occurrences start with a lowercase letter, which drops names;
* is not a plural or articled form of another word (`dreptul`, `marea`), not
  a Roman numeral, and not in `blocklist.txt` (vulgar and offensive words,
  wiki markup).

The 7,776 most frequent words that pass are sorted and numbered.

## Rebuilding

`counts.tsv` holds the word counts from the dump, so the list can be rebuilt
without downloading anything but the dictionary:

```sh
curl -LO https://raw.githubusercontent.com/wooorm/dictionaries/main/dictionaries/ro/index.dic
go run ./scripts/genro -dic index.dic -out internal/wordlist/ro_diceware.txt -min 10 -share 0.5
```

To redo the counts, remove `counts.tsv` and add `-wiki` with the dump
(about 790 MB; the counting takes a few minutes):

```sh
curl -LO https://dumps.wikimedia.org/rowiki/20261101/rowiki-20261101-pages-articles.xml.bz2
go run ./scripts/genro -wiki rowiki-20261101-pages-articles.xml.bz2 -dic index.dic -out internal/wordlist/ro_diceware.txt -min 10 -share 0.5
```

The dictionary is `dictionaries/ro/index.dic` of
[wooorm/dictionaries](https://github.com/wooorm/dictionaries), a packaging of
rospell 3.3.10. Licences: see `NOTICE`.

Changing the list changes the passphrases that a given dice roll gives, so
treat a rebuild as a breaking change of the Romanian and Mixed modes.
