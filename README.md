# five-letter-word-cliques

I wrote this little program in a day with the purpose of finding a list of 5 words with unique characters, from a wordlist, with the use of Go's concurrency patterns.

See https://www.youtube.com/watch?v=_-AfhLQfb6w

### Useful Regex

With the `-output-list` the program will output the list of words without repeating letters and in alphabetical order. If you save this list in a file, you can use the following regex by replacing "chimp" and "flogs" to find all words that could follow.

`/^[^chimpflogs\d\s]{5}\b/`
