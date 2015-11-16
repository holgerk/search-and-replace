[![Build Status](https://travis-ci.org/holgerk/search-and-replace.svg)](https://travis-ci.org/holgerk/search-and-replace)
[![Coverage](http://gocover.io/_badge/github.com/holgerk/search-and-replace?0)](http://gocover.io/github.com/holgerk/search-and-replace)

# Search And Replace

## Features
- search and replace a string in the current directory
- regular expressions in search string and backreferences in replace string

## Usage
```
Usage:
  search-and-replace [OPTIONS] Search Replace

Application Options:
  -d, --dry-run      Do not change anything
  -r, --regexp       Treat search string as regular expression
  -v, --verbose      Show verbose debug information
  -i, --interactive  Confirm every replacement

Help Options:
  -h, --help         Show this help message

Arguments:
  Search
  Replace
```

## Demo (Interactive Mode)
![demo-interactive-mode](https://cloud.githubusercontent.com/assets/1426236/11192315/c7ed5c66-8ca0-11e5-8d8f-46ec8f18d6cd.gif)

## Examples
```
search-and-replace -r "(ba+r)(fo+)" "${2}${1}"
```

## TODO
- ignore files from .gitignore
