//usr/bin/env go run $0 "$@"; exit
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

const chunk_size int = 32 * 4096

type alg_sum struct {
    alg, cksum string
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
    // make go 1.0 happy:
    return sha1.New()
}


func chan_to_hash(ic chan byte, alg string, output_ch chan alg_sum) {
    hash_func := new_hash(alg)
    i := 0
    cnt := 0
    data := make ([]byte, chunk_size)

    for nibble := range ic {
        data[i] = nibble
        i++
        if i == chunk_size {
            hash_func.Write(data)
            i = 0
        }
        //if cnt % 1024 == 0 {
        //    fmt.Printf("chan_to_hash(): Wrote nibble %d\n", cnt)
        //}
        cnt++
    }
    
    output_ch <- alg_sum{alg, hex.EncodeToString(hash_func.Sum(nil))}
}

func hash_chan(alg string, output_ch chan alg_sum) chan byte {
    input_chan := make(chan byte, chunk_size)
    go chan_to_hash(input_chan, alg, output_ch)
    return input_chan
}

func read_fan(input_file *os.File, alg_list []string,
        output_ch chan alg_sum) {
    // create a buffer to keep chunks that are read
    data := make([]byte, chunk_size)

    // prepare a slice of channels for
    // (chunk_size)-chunks:
    input_channels := make([]chan byte, len(alg_list))
    //fmt.Printf("read_fan(): Made slice input_channels of len %d\n", len(input_channels))

    for i, alg := range alg_list {
        //fmt.Printf("read_fan(): calling hash_chan for alg %s\n", alg)
        input_channels[i] = hash_chan(alg, output_ch)
    }
    cnt := 0
    for {
        // read chunks from file:
        num_bytes, err := input_file.Read(data)
        //fmt.Printf("read_fan(): read %d bytes\n", num_bytes)

        // panic on any error != io.EOF
        if err != nil && err != io.EOF { panic(err) }

        // break loop if no more bytes:
        if num_bytes == 0 { break }

        // write data read to channel:
        //if cnt % 1024 == 0 {
        //    fmt.Printf("read_fan(): Sending chunk %d\n", cnt)
        //}
        for _, input_chan := range input_channels {
            //fmt.Printf("read_fan(): sending nibbles to input_channels[%d]\n", i)
            for nibble := range data {
                input_chan <- byte(nibble)
            }
        }
        cnt++
    }
    for _, input_chan := range input_channels {
        close(input_chan)
    }
}

func main() {   
    // filename -> algorithm -> checksum
    output_map :=  map[string]map[string]string{}
    alg_list := []string{"SHA256", "SHA1", "MD5"}
    // put filenames in 1st level of keys:
    for i := 0; i < len(os.Args) -1; i++ {
         output_map[os.Args[i+1]] = make(map[string]string)
    }

    for filename, alg_sum_map := range output_map {
        // open file, exit on error:
        input_file, err := os.Open(filename)
        if err != nil {
            fmt.Println(err)
            return
        }
        //else {
        //    fmt.Printf(" * Opened %s\n", filename)
        //}
        // close on EOF I guess?
        defer input_file.Close()

        output_ch := make(chan alg_sum, len(alg_list))
        go read_fan(input_file, alg_list, output_ch)

        for i := 0; i < len(alg_list); i++ {
              result := <- output_ch
              fmt.Printf("%s-checksum of %s:\n\t%s\n",
                    result.alg, filename, result.cksum)
              alg_sum_map[result.alg] = result.cksum
        }
        close(output_ch)
    }

    for filename, alg_sum_map := range output_map {
        fmt.Printf("%s: \n", filename)
        for alg, cksum := range alg_sum_map {
            fmt.Printf(" - %s: %s\n", alg, cksum)
        }
    }   
}
