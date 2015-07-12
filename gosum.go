/*  (c) 2015 0xf10e@fsfe.org
    Licensed under Apache License 2.0 
    
    Accepts a list of files and computes
    various checksums for every file w/o
    reading the file multiple times

*/
package main

import (
    //"crypto/md5"
    //"crypto/sha1"
    "crypto/sha256"
    "encoding/hex"
    "fmt"
    "io"
    "os"
)

func read_file() {
    /*  reads from file, duplicates into 
        channels feeding into calculation
        go-routines */
}


// func calc(chan input, string alg) {
//     /*  reads chunks of data from
//         input channel and feeds then
//         into specified cksum-algorithm */
// }


func main() {   
    // filename -> algorithm -> checksum
    output_map :=  map[string]map[string]string{}

    for i := 0; i < len(os.Args) -1; i++ {
         output_map[os.Args[i+1]] = make(map[string]string)
    }
    fmt.Println("Files:")
    for key, _ := range output_map {
        fmt.Printf("- %s\n", key)
    }

    // open file, exit on error:
    input_file, err := os.Open(os.Args[1])
    if err != nil {
        fmt.Println(err)
        return
    }
    // close on EOF I guess?
    defer input_file.Close()

    // create a buffer to keep chunks that are read
    data := make([]byte, 16)

    // initialize hash-func:
    hash := sha256.New()

    for {
        // read chunks from file:
        num_bytes, err := input_file.Read(data)
        // panic on any error != io.EOF
        if err != nil && err != io.EOF { panic(err) }
        // break loop if no more bytes:
        if num_bytes == 0 { break }       
        // write data read to hashing function:
        hash.Write(data)
    }
    fmt.Println(hex.EncodeToString(hash.Sum(nil)))
}
