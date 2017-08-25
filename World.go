package main;

// import "log"
// import "encoding/json"
// import "net/http"
// import "time"
// import "fmt"
import "io/ioutil"
// import "crypto/sha256"
// import "crypto/sha512"
import "crypto/md5"
// import "strings"
// import "errors"
// import "strconv"
import "sync"
// import "crypto/rand"
// import "os"


type World struct{
    seed []byte
    cache map[[2]int64][2]byte // (x,y) -> (field kind, abundance)
    cache_rwlock *sync.RWMutex
}

func new_World() World{
    seed, err:=ioutil.ReadFile(world_seed_file)
    if err!=nil{
        panic(err.Error())
    }
    return World{seed, make(map[[2]int64][2]byte), &sync.RWMutex{}}
}

func (w *World) get_fields_static_data(x,y int64) (byte, byte){
    // If it is cached, use cached return value
    w.cache_rwlock.RLock()
    if field_static_data,ok:=w.cache[[2]int64{x,y}];ok{
        w.cache_rwlock.RUnlock()
        return field_static_data[0], field_static_data[1]
    }
    w.cache_rwlock.RUnlock()

    // Actually calculate return value
    kind,quantity:=func() (byte,byte){
        h := md5.New()
        h.Write(w.seed)
        h.Write([]byte{byte(x&0xFF), byte((x>>8)&0xFF), byte(y&0xFF), byte((y>>8)&0xFF)})
        s:=h.Sum(nil)

        quantity:=(s[1]&0x3F)+36
        if s[0]<193{
            return 0, quantity
        } else if s[0]<193+36{
            return 1, quantity
        } else if s[0]<193+36+18{
            return 2, quantity
        } else{ //  if s[0]<193+36+18+9
            return 3, quantity
        }
    }()

    //If return value was cached in the meantime use cached return value, otherwise cache calculated return value and return it
    w.cache_rwlock.Lock()
    if field_static_data,ok:=w.cache[[2]int64{x,y}];ok{
        w.cache_rwlock.Unlock()
        return field_static_data[0], field_static_data[1]
    } else{
        field_static_data:=[2]byte{kind,quantity}
        w.cache[[2]int64{x,y}]=field_static_data
        w.cache_rwlock.Unlock()
        return field_static_data[0], field_static_data[1]
    }
    // w.cache_rwlock.Unlock()
}