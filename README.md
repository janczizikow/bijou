# bijou 💎

https://github.com/prettier/prettier/tree/0.0.1
https://github.com/prettier/prettier/tree/1.0.0
https://github.com/estree/estree


TODOS:
- [ ] lexer
  - [ ] handle all tokens?
    - [ ] private identifiers (eg.: #foo)
    - [ ] TypeScript decorators (eg.: @sealed)
    - [ ] comments :(
  - [x] string literals
    - [x] escaped chars?
    - [ ] template literals?
  - [x] numeric literals
    - [x] 1_000_000
    - [x] bigint (eg.: 123456789123456789n)
  - [ ] line tracking?
  - [x] error handling
    - [ ] track lines?
    - [ ] show errors verbosely with ^
- [ ] parser
  - [ ] parsing bindings like ({ msg, err })
  - [ ] figure out how to deal with operator precedence
- [x] printer
  - [ ] bugs

## Usage

