package main

import (
    "encoding/csv"
    "flag"
    "fmt"
    "io"
    "io/ioutil"
    "os"
    "strconv"
    "strings"
    "unicode/utf8"
)

var separator = flag.String("separator", ",", "Single character to be used as a separator between fields")
var fieldNumsRaw = flag.String("fieldNums", "", "Comma-separated list of field indexes (starting at 0) to print to the command line")

var noPrintRealCSV = flag.Bool("noPrintCSV", false, "Program defaults to printing valid, quoted, well-formatted CSV. If this flag is supplied, output is returned as a string joined by outJoinStr. noPrintCSV is assumed to imply you want to pass the output to naive tools like cut or awk.")
var outputJoiner = flag.String("outJoinStr", ",", "Separator to use when printing multiple columns in your output. Only valid if outputting something meant to be passed to cut/awk, and not a properly-formatted, quoted CSV file.")

func main() {
    flag.Parse()

    var fieldNums []int

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

    // TODO: Make this stream from stdin, and also stream from a file
    bytes, err := ioutil.ReadAll(os.Stdin)
    if err != nil {
        panic(err)
    }

    lines := strings.Split(string(bytes), "\n")

    csvWriter := csv.NewWriter(os.Stdout)
    csvWriter.Comma = getSeparator(*outputJoiner)

    for _, line := range lines {
        fields, err := processLine(line)
        if err != nil {
            if err == io.EOF {
                break
            } else {
                panic(err)
            }
        }

        var toPrint []string
        if *fieldNumsRaw == "" {
            for i, _ := range fields {
                toPrint = append(toPrint, fields[i])
            }
        } else {
            for _, num := range fieldNums {
                toPrint = append(toPrint, fields[num])
            }
        }

        if *noPrintRealCSV == false {
            csvWriter.Write(toPrint)
        } else {
            fmt.Println(strings.Join(toPrint, *outputJoiner))
        }
    }
    if *noPrintRealCSV == false {
        csvWriter.Flush()
    }
}

func processLine(line string) ([]string, error) {
    strReader := strings.NewReader(line)
    csvReader := csv.NewReader(strReader)

    sepString := *separator
    _ = utf8.DecodeRuneInString
    _ = sepString

    csvReader.Comma = getSeparator(sepString)

    fields, err := csvReader.Read()
    if err != nil {
        if err == io.EOF {
            return nil, io.EOF
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
        if err.Error() == "invalid syntax" {  // Single quote used as separator. No idea why someone would want this, but it doesn't hurt to support it
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
