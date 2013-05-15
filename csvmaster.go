package main

/*
    TODO:
        - Read stdin line by line instead of all at once ( http://stackoverflow.com/questions/12363030/read-from-initial-stdin-in-go )
        - Make it use the CSV writer to write out the stuff instead of println
        - Use flag parsing to determine which field(s) to print out
        - Allow writing out the data with a different delimiter than what it came in with
*/

import (
    //"bufio"
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
var outputJoiner = flag.String("outJoinStr", ",", "Separator to use when printing multiple columns in your output. Only valid if outputting something meant to be passed to cut/awk, and not a properly-formatted, quoted CSV file.")

func main() {
    flag.Parse()

    var fieldNums []int

    for _, numStr := range strings.Split(*fieldNumsRaw, ",") {
        numStr := strings.TrimSpace(numStr)
        numInt, err := strconv.Atoi(numStr)
        if err != nil {
            panic(err)
        }
        fieldNums = append(fieldNums, numInt)
    }

    bytes, err := ioutil.ReadAll(os.Stdin)
    if err != nil {
        panic(err)
    }

    lines := strings.Split(string(bytes), "\n")

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
        for _, num := range fieldNums {
            toPrint = append(toPrint, fields[num])
        }
        fmt.Println(strings.Join(toPrint, *outputJoiner))
    }
}

func processLine(line string) ([]string, error) {
    strReader := strings.NewReader(line)
    csvReader := csv.NewReader(strReader)

    sepString := *separator
    /*
    fmt.Println("Separator is", string(sepString[0]))
    fmt.Println("'", rune(sepString[0]), "'")
    fmt.Println("'", string(rune("\t"[0])), "'")
    r, size := utf8.DecodeRuneInString(sepString)
    fmt.Println("'", string(r), "'")
    fmt.Println(size)
    */
    _ = utf8.DecodeRuneInString

    csvReader.Comma = rune(sepString[0])

    fields, err := csvReader.Read()
    if err != nil {
        if err == io.EOF {
            return nil, io.EOF
        } else {
            panic(err)
        }
    }

    return fields, nil
}
