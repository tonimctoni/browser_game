package main;

// import "log"
// import "encoding/json"
// import "net/http"
// import "time"
// import "fmt"
// import "io/ioutil"
import "crypto/sha256"
// import "crypto/sha512"
// import "crypto/md5"
// import "strings"
import "errors"
// import "strconv"
import "sync"
import "crypto/rand"
import "os"


type LoginData struct{
    id uint64
    permissions uint16
    login_username string
    salt [8]byte
    salted_password_hash [32]byte
}

func (l *LoginData) is_password_correct(password string) bool{
    h := sha256.New()
    h.Write([]byte(password))
    h.Write(l.salt[:])
    bs := h.Sum(nil)

    for i:=0; i<32; i++{
        if bs[i]!=l.salted_password_hash[i]{
            return false
        }
    }

    return true
}


// Just for some encapsulation <== program is getting complex
type LoginMap struct{
    login_map map[string]*LoginData // (username -> all login data)
    add_login_rwlock *sync.RWMutex
    filename string
}

// This handles the reading of user login data from disk
func load_LoginMap(filename string) LoginMap{
    f, err:=os.Open(filename)
    if err!=nil{
        panic("Could not open login file")
    }
    defer f.Close()

    login_map:=make(map[string]*LoginData)
    buffer:=make([]byte, 74)
    for{
        n,err:=f.Read(buffer[:])
        if err!=nil{
            if n!=0{
                panic("Incomplete portion of data read from login file")
            }
            break
        }

        login_data:=&LoginData{byte_slice_to_uint64(buffer[0:8]),byte_slice_to_uint16(buffer[8:10]),byte_slice24_to_string(buffer[10:34]),byte_slice8_to_array8(buffer[34:42]),byte_slice32_to_array32(buffer[42:74])}
        _,ok:=login_map[login_data.login_username]
        if ok{
            panic("Duplicate login name (load_LoginMap)")
        }
        login_map[login_data.login_username]=login_data
    }

    return LoginMap{login_map, &sync.RWMutex{}, filename}
}

func (l *LoginMap) get_login_data(login_name string) (*LoginData, bool){
    l.add_login_rwlock.RLock()
    login_data, ok:=l.login_map[login_name]
    l.add_login_rwlock.RUnlock()
    return login_data, ok
}

func (l *LoginMap) add_login(permissions uint16, login_name, password string) (uint64, error){ // returns new ID
    if len(login_name)>24{
        return 0, errors.New("Login_name is too long (len(login_name)>24, add_login)")
    }

    l.add_login_rwlock.Lock()
    defer l.add_login_rwlock.Unlock() 

    // Check if name already exists in map
    _,ok:=l.login_map[login_name]
    if ok{
        // panic("Duplicate login name (add_login)")
        return 0, errors.New("Duplicate login name (add_login)")
    }

    // Get salt
    salt:=[8]byte{}
    _,err:=rand.Read(salt[:])
    if err!=nil{
        // panic("Could not read random data (add_login)")
        return 0, errors.New("Could not read random data (add_login)")
    }

    h := sha256.New()
    h.Write([]byte(password))
    h.Write(salt[:])
    salted_password_hash:=h.Sum(nil)

    f, err:=os.OpenFile(l.filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
    if err!=nil{
        // panic("Could not open file (add_login)")
        return 0, errors.New("Could not open file (add_login)")
    }
    defer f.Close()

    id:=uint64(len(l.login_map)+1)
    f.Write(uint64_to_bytes(id))
    f.Write(uint16_to_bytes(permissions))
    f.Write(string_to_24_bytes(login_name))
    f.Write(salt[:])
    f.Write(salted_password_hash)

    salted_password_hash_array:=[32]byte{}
    copy(salted_password_hash_array[:],salted_password_hash)
    l.login_map[login_name]=&LoginData{id, permissions, login_name, salt, salted_password_hash_array}

    return id, nil
}
