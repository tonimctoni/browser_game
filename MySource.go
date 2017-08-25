package main;

// import "log"
// import "encoding/json"
// import "net/http"
// import "time"
// import "fmt"
// import "io/ioutil"
// import "crypto/sha256"
import "crypto/sha512"
// import "crypto/md5"
// import "strings"
// import "errors"
// import "strconv"
// import "sync"
// import "crypto/rand"
// import "os"

type MySource struct{
    state [64]byte
}

func new_MySource(initial_state [64]byte) MySource{
    return MySource{initial_state}
}

func (m *MySource) Seed(s int64) {
    seed:=uint64(s)
    for i := uint(0); i < 64; i++ {
        m.state[i]=byte(((seed>>i)^(seed>>(64-i-1)))&0xff);
    }
}

func (m *MySource) Int63() int64{
    m.state=sha512.Sum512(m.state[:])
    for m.state[7]&0x80!=0{ //Has two functions: makes sure the last bit is not set (int63 after all), and adds some self shrinking
        m.state=sha512.Sum512(m.state[:])
    }

    return int64(m.state[0]) | (int64(m.state[1]) << 8) | (int64(m.state[2]) << 16) | (int64(m.state[3]) << 24) |
        (int64(m.state[4]) << 32) | (int64(m.state[5]) << 40) | (int64(m.state[6]) << 48) | (int64(m.state[7]) << 56)
}