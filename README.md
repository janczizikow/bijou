# bijou

`bijou` is an experimental (unfinished 😅) CLI tool built to explore how code formatters work under the hood.
It parses JavaScript source code into an AST and aims to print formatted code using [Philip Wadler’s pretty-printing algorithm](https://homepages.inf.ed.ac.uk/wadler/papers/prettier/prettier.pdf) (the same algorithm behind Prettier).

Pipeline:

source -> tokens -> parse (AST) -> AST to DOC -> output

_Most formatting options aren't respected yet, and the printer component is due for a rewrite._

## Why?

Mostly to get more hands-on experience with Go.
With TypeScript 7.0 already written in go, this probably no longer makes sense.
There are already non-JS tools that can format JavaScript/TypeScript, but I wanted to understand how do such tools work.

## Usage

### 1. Using the Go Binary

```bash
# Build the binary
make build
# Create a test file
echo 'function greet(name) {console.log("Hello " + name)}' > test.js

./bin/bijou test.js # prints to stdout
./bin/bijou -w test.js # edit the file in-place
```

### 2. Using via Node.js / npm

Could also execute it from node using `cli.js` / npm package:

```bash
make build
npm pack
npm install bijou-0.0.0-dev.tgz
npx bijou test.js
```

## roadmap

realistically this might never happen:

- [ ] rewrite the printer component using Wadler's pretty-printer
- [ ] parse config flies to load formatting options
- [ ] add support for missing JS syntax lexer/parser
- [ ] Add JSX and TypeScript syntax?
