package main

import (
    "bufio"
    "encoding/csv"
    "flag"
    "fmt"
    "io"
    "os"
    "strconv"
    "strings"
    "syscall"
)

const NILCOMMENTRUNE = "TOTALLYNOTACOMMENTCHAR"
const VERSION = "1.1"

// TODO: Add support for specifying fields by field header name instead of just number
// TODO: check ReadRuneFromString instead of existing technique
// TODO: Publish binaries
// TODO: Include test suite
// TODO: Accept filename without --filename flag

var inSep string
var outSep string
var commentRune string

var filename string

var fieldNumsRaw string
var noRFC = flag.Bool("no-rfc", false, "Program defaults to printing RFC 4180-compliant, quoted, well-formatted CSV. If this flag is supplied, output is returned as a string naively joined by --out-sep. --no-rfc is assumed to imply you want to pass the output to naive tools like cut or awk, and in that case, it is recommended that you select an --out-sep that is unlikely to be in youc content, such as a pipe or a backtick.")

var printVersion bool

func init() {
    const (
        filenameFlagLong = "file"
        filenameFlagShort = "f"
        filenameDesc = "File to read from. If not specified, program reads from stdin."

        fieldNumsRawFlagLong = "fields"
        fieldNumsRawFlagShort = "F"
        fieldNumsRawDesc =  "Comma-separated list of field indexes (starting at 0) to print to the command line"

        inSepFlagLong = "in-sep"
        inSepFlagShort = "i"
        inSepDesc = "Single character field separator used by your input"
        inSepDefault = ","

        outSepFlagLong = "out-sep"
        outSepFlagShort = "o"
        outSepDesc = "Single-character field separator to use when printing multiple columns in your output. Only valid if outputting something meant to be passed to cut/awk, and not a properly-formatted, quoted CSV file."
        outSepDefault = ","

        commentRuneFlagLong = "comment-char"
        commentRuneFlagShort = "c"
        commentRuneDesc = "Single-character field separator to use when printing multiple columns in your output. Only valid if outputting something meant to be passed to cut/awk, and not a properly-formatted, quoted CSV file."

        printVersionFlagLong = "version"
        printVersionFlagShort = "v"
        printVersionDesc = "Print program version"
        printVersionDefault = false
    )

    flag.StringVar(&filename, filenameFlagLong, "", filenameDesc)
    flag.StringVar(&filename, filenameFlagShort, "", filenameDesc)

    flag.StringVar(&fieldNumsRaw, fieldNumsRawFlagLong, "", fieldNumsRawDesc)
    flag.StringVar(&fieldNumsRaw, fieldNumsRawFlagShort, "", fieldNumsRawDesc)

    flag.StringVar(&inSep, inSepFlagLong, inSepDefault, inSepDesc)
    flag.StringVar(&inSep, inSepFlagShort, inSepDefault, inSepDesc)

    flag.StringVar(&outSep, outSepFlagLong, outSepDefault, outSepDesc)
    flag.StringVar(&outSep, outSepFlagShort, outSepDefault, outSepDesc)

    flag.StringVar(&commentRune, commentRuneFlagLong, NILCOMMENTRUNE, commentRuneDesc)
    flag.StringVar(&commentRune, commentRuneFlagShort, NILCOMMENTRUNE, commentRuneDesc)

    flag.BoolVar(&printVersion, printVersionFlagLong, printVersionDefault, printVersionDesc)
    flag.BoolVar(&printVersion, printVersionFlagShort, printVersionDefault, printVersionDesc)
}

func main() {
    flag.Parse()

    if printVersion == true {
        fmt.Fprintf(os.Stdout, "csvmaster version %s, made by spiffytech. http://github.com/spiffytech/csvmaster\n", VERSION)
        syscall.Exit(0)
    }

    var fieldNums []int

    // Identify the numbers you want to print out
    fieldNumsRaw = strings.Trim(fieldNumsRaw, ",")
    if fieldNumsRaw != "" {
        for _, numStr := range strings.Split(fieldNumsRaw, ",") {
            numStr := strings.TrimSpace(numStr)
            numInt, err := strconv.Atoi(numStr)
            if err != nil {
                panic(err)
            }
            fieldNums = append(fieldNums, numInt)
        }
    }

    var reader *bufio.Reader
    if filename == "" {
        reader = bufio.NewReader(os.Stdin)
    } else {
        f, err := os.Open(filename)
        if err != nil {
            panic(err)
        }

        reader = bufio.NewReader(f)
    }

    csvReader := csv.NewReader(reader)
    csvReader.LazyQuotes = true
    csvReader.TrailingComma = true
    csvReader.Comma = getSeparator(inSep)
    csvReader.FieldsPerRecord = -1
    if commentRune != NILCOMMENTRUNE {
        csvReader.Comment = ([]rune(commentRune))[0]
    }

    csvWriter := csv.NewWriter(os.Stdout)
    csvWriter.Comma = getSeparator(outSep)

    for {
        fields, err := csvReader.Read()
        if err != nil {
            if err == io.EOF {
                break
            }
            panic(err)
        }

        var toPrint []string
        if fieldNumsRaw == "" {  // Print all fields
            toPrint = fields
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
            fmt.Println(strings.Join(toPrint, outSep))
        }
    }
    if *noRFC == false {
        csvWriter.Flush()
    }
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
