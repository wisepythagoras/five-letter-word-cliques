# five-letter-word-cliques

I wrote this little program in a day with the purpose of finding a list of 5 words with unique characters, from a wordlist, with the use of Go's concurrency patterns. I used the [backtracking](https://en.wikipedia.org/wiki/Backtracking) algorithm, which isn't the most time-efficient way (`O(4^n)`), but it gets the job done in a just over 5 minutes for a list of ~5600 words (`-find-all`).

See https://www.youtube.com/watch?v=_-AfhLQfb6w

### Useful Regex

With the `-output-list` the program will output the list of words without repeating letters and in alphabetical order. If you save this list in a file, you can use the following regex by replacing "chimp" and "flogs" to find all words that could follow.

`/^[^chimpflogs\d\s]{5}\b/`
