package main

import (
    "bufio"
    "encoding/csv"
    "errors"
    "flag"
    "fmt"
    "io"
    "os"
    "strconv"
    "strings"
)

const nilCommentRune = "TOTALLYNOTACOMMENTCHAR"

// TODO: Do I need to have processLine() at all? Can I just use csvReader.Read on the buffer?
// TODO: Add support for specifying fields by field header name instead of just number
// TODO: check ReadRuneFromString instead of existing technique

var inSep = flag.String("in-sep", ",", "Single character field separator used by your input")
var outSep = flag.String("out-sep", ",", "Single-character field separator to use when printing multiple columns in your output. Only valid if outputting something meant to be passed to cut/awk, and not a properly-formatted, quoted CSV file.")
var commentRune = flag.String("comment-char", nilCommentRune, "Single-character field separator to use when printing multiple columns in your output. Only valid if outputting something meant to be passed to cut/awk, and not a properly-formatted, quoted CSV file.")

var filename = flag.String("filename", "", "File to read from. If not specified, program reads from stdin.")

var fieldNumsRaw = flag.String("field-nums", "", "Comma-separated list of field indexes (starting at 0) to print to the command line")
var noRFC = flag.Bool("no-rfc", false, "Program defaults to printing RFC 4180-compliant, quoted, well-formatted CSV. If this flag is supplied, output is returned as a string naively joined by --out-sep. --no-rfc is assumed to imply you want to pass the output to naive tools like cut or awk, and in that case, it is recommended that you select an --out-sep that is unlikely to be in youc content, such as a pipe or a backtick.")

func main() {
    flag.Parse()

    var fieldNums []int

    // Identify the numbers you want to print out
    *fieldNumsRaw = strings.Trim(*fieldNumsRaw, ",")
    if *fieldNumsRaw != "" {
        for _, numStr := range strings.Split(*fieldNumsRaw, ",") {
            numStr := strings.TrimSpace(numStr)
            numInt, err := strconv.Atoi(numStr)
            if err != nil {
                panic(err)
            }
            fieldNums = append(fieldNums, numInt)
        }
    }

    csvWriter := csv.NewWriter(os.Stdout)
    csvWriter.Comma = getSeparator(*outSep)

    var reader *bufio.Reader
    if *filename == "" {
        reader = bufio.NewReader(os.Stdin)
    } else {
        f, err := os.Open(*filename)
        if err != nil {
            panic(err)
        }

        reader = bufio.NewReader(f)
    }

    for {
        line, err := reader.ReadString('\n')
        if err != nil {
            if err == io.EOF {
                break
            }
            panic(err)
        }

        if strings.TrimSpace(line) == "" {
            fmt.Println(line)
            continue
        }

        fields, err := processLine(line, false)
        if err != nil {
            if err == io.EOF {
                continue  // Since it's only one line, and not the whole file, it's a bogus EOF that happens when e.g. your line starts with a comment. File EOF is handled above.
            } else if err.Error() == "Incomplete value" {  // Handles values containing newlines
                moreLine, err := reader.ReadString('\n')
                if err != nil {
                    panic(err)
                }
                line += moreLine
                //moreLine, err = reader.ReadString('\n')
                //if err != nil {
                //    panic(err)
                //}
                //line += moreLine
                fields, err = processLine(line, false)
            } else {
                panic(err)
            }
        }

        var toPrint []string
        if *fieldNumsRaw == "" {  // Print all fields
            for i, _ := range fields {
                toPrint = append(toPrint, fields[i])
            }
        } else {
            for _, num := range fieldNums {
                if num > len(fields) - 1 {
                    toPrint = append(toPrint, "")  // Append _something_, so printing columns out of order preserves the column index:value mapping in all columns. I.e., --field-nums=1,2,0 on a 2-column line will print "a,,b" so you always know your knew third column has certain content
                } else {
                    toPrint = append(toPrint, fields[num])
                }
            }
        }

        if *noRFC == false {
            csvWriter.Write(toPrint)
        } else {
            fmt.Println(strings.Join(toPrint, *outSep))
        }
    }
    if *noRFC == false {
        csvWriter.Flush()
    }
}

func processLine(line string, lazyQuotes bool) ([]string, error) {
    strReader := strings.NewReader(line)
    csvReader := csv.NewReader(strReader)
    csvReader.LazyQuotes = lazyQuotes
    csvReader.TrailingComma = true
    if *commentRune != nilCommentRune {
        csvReader.Comment = ([]rune(*commentRune))[0]
    }

    sepString := *inSep
    csvReader.Comma = getSeparator(sepString)

    fields, err := csvReader.Read()
    if err != nil {
        if err == io.EOF {
            return nil, io.EOF
        } else if strings.Contains(err.Error(), "in non-quoted-field") {  // Field isn't quoted, but contains quotes
            return processLine(line, true)
        } else if strings.Contains(err.Error(), `extraneous " in field`) {  // Field is quoted and ends with a newline that's part of the value
            return nil, errors.New("Incomplete value")
        } else {
            fmt.Println("Error in the following line:")
            fmt.Println(line)
            panic(err)
        }
    }

    return fields, nil
}

func getSeparator(sepString string) (sepRune rune) {
    sepString = `'` + sepString + `'`
    sepRunes, err := strconv.Unquote(sepString)
    if err != nil {
        if err.Error() == "invalid syntax" {  // Single quote was used as separator. No idea why someone would want this, but it doesn't hurt to support it
            sepString = `"` + sepString + `"`
            sepRunes, err = strconv.Unquote(sepString)
            if err != nil {
                panic(err)
            }

        } else {
            panic(err)
        }
    }
    sepRune = ([]rune(sepRunes))[0]

    return sepRune
}
