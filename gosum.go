/*  (c) 2015 0xf10e@fsfe.org
    Licensed under Apache License 2.0 
    
    Accepts a list of files and computes
    various checksums for every file w/o
    reading the file multiple times

*/
package main

import (
    "crypto/md5"
    "crypto/sha1"
    "crypto/sha256"
    "encoding/hex"
    "fmt"
    "hash"
    "io"
    "os"
)


type file_alg_sum struct {
    filename, alg, cksum string
}

func new_hash(alg string) hash.Hash {
    switch alg {
    case "SHA256":
        return sha256.New()
    case "MD5":
        return md5.New()
    case "SHA1":
        return sha1.New()
    default:
        panic ("Unknown algorithm!")
    }
}


func chan_to_hash(ic chan byte, file string, alg string, oc chan file_alg_sum) {
    hash_func := new_hash(alg)
    i := 0
    cnt := 0
    data := make ([]byte, 16)

    for nibble := range ic {
        data[i] = nibble
        i++
        if i == 16 {
            hash_func.Write(data)
            i = 0
        }
        if cnt % 1024 == 0 {
            fmt.Printf("chan_to_hash(): Wrote nibble %d\n", cnt)
        }
        cnt++
    }
    
    oc <- file_alg_sum{file, alg, hex.EncodeToString(hash_func.Sum(nil))}
}

func read_routine(input_file *os.File, ic chan byte) {
    // create a buffer to keep chunks that are read
    data := make([]byte, 16)

    cnt := 0
    for {
        // read chunks from file:
        num_bytes, err := input_file.Read(data)
        // panic on any error != io.EOF
        if err != nil && err != io.EOF { panic(err) }
        // break loop if no more bytes:
        if num_bytes == 0 { break }
        // write data read to channel:
        if cnt % 1024 == 0 {
            fmt.Printf("main(): Sending chunk %d\n", cnt)
        }
        for nibble := range data {
            ic <- byte(nibble)
        }
        cnt++
    }
    close(ic)
}

func main() {   
    // filename -> algorithm -> checksum
    output_map :=  map[string]map[string]string{}
    algorithms := []string{"SHA256", "SHA1", "MD5"}
    // put filenames in 1st level of keys:
    for i := 0; i < len(os.Args) -1; i++ {
         output_map[os.Args[i+1]] = make(map[string]string)
    }

    for filename, alg_sum := range output_map {
        // open file, exit on error:
        input_file, err := os.Open(filename)
        if err != nil {
            fmt.Println(err)
            return
        } else {
            fmt.Printf(" * Opened %s\n", filename)
        }
        // close on EOF I guess?
        defer input_file.Close()

        ic := make(chan byte, 16)
        oc := make(chan file_alg_sum, 1)
        go read_routine(input_file, ic)
        for _, alg := range algorithms {
            go chan_to_hash(ic, filename, alg, oc)
            break
        }
        // for len(ALGORITHMS)
        result := <- oc
        //fmt.Printf("oc returns %s for %s of\n%s", 
        //        result.alg, filename, result.cksum)
        alg_sum[result.alg] = result.cksum
    }

    for filename, alg_sum := range output_map {
        fmt.Printf("%s: \n", filename)
        for alg, cksum := range alg_sum {
            fmt.Printf(" - %s: %s\n", alg, cksum)
        }
    }   
}
