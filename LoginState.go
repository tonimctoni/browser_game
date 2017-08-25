package main;

// import "log"
// import "encoding/json"
// import "net/http"
import "time"
// import "fmt"
// import "io/ioutil"
// import "crypto/sha256"
// import "crypto/sha512"
// import "crypto/md5"
// import "strings"
// import "errors"
// import "strconv"
import "sync"
// import "crypto/rand"
// import "os"


type LoginState struct{
    player_data *PlayerData
    token string
    ip string
    expire_date time.Time
}

type LoginStateMap struct{
    login_state_map map[uint64]LoginState // (user_id -> his login state)
    login_state_map_rwlock *sync.RWMutex
}

func new_LoginStateMap() LoginStateMap{
    return LoginStateMap{make(map[uint64]LoginState), &sync.RWMutex{}}
}

func (l *LoginStateMap) get_login_state(id uint64) (LoginState, bool){
    l.login_state_map_rwlock.RLock()
    login_state, ok:=l.login_state_map[id]
    l.login_state_map_rwlock.RUnlock()

    return login_state, ok
}

func (l *LoginStateMap) set_login_state(id uint64, login_state LoginState){
    l.login_state_map_rwlock.Lock()
    l.login_state_map[id]=login_state
    l.login_state_map_rwlock.Unlock()
}