//usr/bin/env go run $0 "$@"; exit
/*  (c) 2015 0xf10e@fsfe.org
    Licensed under Apache License 2.0 
    
    Accepts a list of files and computes
    various checksums for every file w/o
    reading the file multiple times

*/
package main

import (
    "bytecount"
    "crypto/md5"
    "crypto/sha1"
    "crypto/sha256"
    "encoding/hex"
    // for profiling:
    "flag"
    "fmt"
    "hash"
    "hash/crc32"
    "io"
    "os"
    "runtime"
    // for profiling:
    "runtime/pprof"
)

var alg_list = []string{"SHA256", "SHA1", "MD5", "CRC", "bytecount"}
var num_threads = flag.Int("t", 1, "sets GOMAXPROCS")
var chunk_size int
func init() {
    flag.IntVar(&chunk_size, "s", 128 * 1024, "set chunk_size")
}

// for profiling:
var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

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
    case "CRC":
        return crc32.NewIEEE()
    case "bytecount":
        return bytecount.New()
    default:
        panic ("Unknown algorithm!")
    }
    // make go 1.0 happy:
    return sha1.New()
}


func chan_to_hash(ic chan int, alg string, buffer []byte, write_ok chan int,
        output_ch chan alg_sum) {
    // get the method for calculation
    // given hash:
    hash_func := new_hash(alg)

    // everytime we get the number
    // of new input bytes we add those
    // to the hash_func:
    for num_bytes := range ic {
        hash_func.Write(buffer[0:num_bytes])
        // write_ok is chan int in case we'll
        // use more then one buffer later:
        write_ok <- 0
    }
    output_ch <- alg_sum{alg, hex.EncodeToString(hash_func.Sum(nil))}
}

func hash_chan(alg string, buffer []byte, write_ok chan int,
        output_ch chan alg_sum) chan int {
    input_chan := make(chan int, 8)
    go chan_to_hash(input_chan, alg, buffer, write_ok, output_ch)
    return input_chan
}

func read_fan(input_file *os.File, alg_list []string,
        output_ch chan alg_sum) {
    // create a buffer to keep chunks that are read
    data := make([]byte, chunk_size)

    // prepare a slice of channels for telling
    // routines how many bytes to read:
    input_channels := make([]chan int, len(alg_list))

    // open the channel routines will use to
    // tell when refilling the buffer is OK:
    write_ok := make(chan int, len(alg_list))
    // (write_ok is chan int in case we'll
    // use more then one buffer later)

    for i, alg := range alg_list {
        //fmt.Printf("read_fan(): calling hash_chan for alg %s\n", alg)
        input_channels[i] = hash_chan(alg, data, write_ok, output_ch)
    }
    for {
        // read chunks from file:
        num_bytes, err := input_file.Read(data)
        //fmt.Printf("read_fan(): read %d bytes\n", num_bytes)

        // panic on any error != io.EOF
        if err != nil && err != io.EOF { panic(err) }

        // break loop if no more bytes:
        if num_bytes == 0 {
            for _, input_chan := range input_channels {
                close(input_chan)
            }
            break
        }

        // write data read to channel:
        //if cnt % 1024 == 0 {
        //    fmt.Printf("read_fan(): Sending chunk %d\n", cnt)
        //}
        for _, input_chan := range input_channels {
            input_chan <- num_bytes
        }
        for i := 0; i < len(alg_list); i++ {
            // wait for every routine giving us
            // an OK for refilling the buffer
            <- write_ok
        }
    }
    // expicitly closing the input file:
    input_file.Close()
}

func main() {
    flag.Parse()
    runtime.GOMAXPROCS(*num_threads)

    if *cpuprofile != "" {
        f, err := os.Create(*cpuprofile)
        if err != nil {
            fmt.Println(err)
            return
        }
        pprof.StartCPUProfile(f)
        defer pprof.StopCPUProfile()
    }
    // filename -> algorithm -> checksum
    output_map :=  map[string]map[string]string{}
    // put filenames in 1st level of keys:
    for _, file := range flag.Args() {
         output_map[file] = make(map[string]string)
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
              alg_sum_map[result.alg] = result.cksum
        }
        close(output_ch)
    }

    for filename, alg_sum_map := range output_map {
        for alg, cksum := range alg_sum_map {
            fmt.Printf("%s (%s) = %s\n", alg, filename, cksum)
        }
    }   
}
