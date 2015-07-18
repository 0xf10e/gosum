package bytecount

//import "fmt"

// behaupten, dass immer 4 Byte rauskommen:
const Size = 4

type digest struct {
    counter uint64
}

func (bc digest) BlockSize() int {
    // every single byte counts!
    return 1
}

func New() *digest {
    return &digest{0}
}

func (bc *digest) Reset() {
    bc.counter = 0
}

func (bc *digest) Size() int {
    // an uint64 is 64bit = 8 bytes
    return 8
}

func (bc *digest) Write(p []byte) (int, error) {
    n := len(p)
    //cnt_old := bc.counter
    bc.counter = bc.counter + uint64(n)
    //fmt.Printf(" * Added %d to %d -> %d\n", n, cnt_old, bc.counter)
    return n, nil
}

func (bc *digest) Sum(b []byte) []byte{
    bc.Write(b)
    ret := []byte {0,0,0,0}
    ret[0] = byte(bc.counter >> 24)
    ret[1] = byte(bc.counter >> 16)
    ret[2] = byte(bc.counter >> 8)
    ret[3] = byte(bc.counter)
    //fmt.Printf(" * Resulting bc.counter = %d\n", bc.counter)
    return ret
}
