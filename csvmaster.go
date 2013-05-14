package main

import (
    //"bufio"
    "encoding/csv"
    "fmt"
    "io"
    "io/ioutil"
    "os"
    "strings"
)

func main() {

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
        fmt.Println(fields[1])
    }
}

func processLine(line string) ([]string, error) {
    strReader := strings.NewReader(line)
    csvReader := csv.NewReader(strReader)

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
