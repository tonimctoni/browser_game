package main;

import "log"
import "encoding/json"
import "net/http"
import "time"
import "fmt"
import "io/ioutil"
import "crypto/sha256"
import "crypto/sha512"
import "crypto/md5"
import "strings"
import "errors"
import "strconv"
import "sync"
import "crypto/rand"

const(
    login_lifetime time.Duration = 2*time.Hour
    available_files_dir = "files"
    available_files_address = "files"
    world_seed_file = "data/world_seed"
    login_data_file = "data/login_data.json"
    player_data_file = "data/player_data.json"
    login_file = "files/login.html"
    home_file = "elm_home/index.html"
    login_address = "/login"
    unlogin_address = "/unlogin"
    get_name_address = "/get_name"
    get_data_address = "/get_data"
    get_world_address = "/get_world"
    root_address = "/"
    home_address = "/home"
    cookie_name = "login_token"

    // time_slot_size = 12*time.Hour
    // steps_received_per_time_slot = 20
    // beginning_of_time = //something like time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
)

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



type LoginData struct{
    Id uint64
    Login_username string
    Salted_password_hash string
    Salt string
}

type PlayerData struct{
    Id uint64
    Name string
    Pos_x int64
    Pos_y int64
    Resource_A int64
    Resource_B int64
    Resource_C int64
    Available_steps int64
    Last_time_gotten_steps int64 // number of time slots elapsed since beginning
}

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



// Just for some encapsulation <== program is getting complex
type LoginMap struct{
    login_map map[string]*LoginData // (username -> all login data)
}

// This handles the reading of user login data from disk
func load_LoginMap() LoginMap{
    // Read whole file
    content, err:=ioutil.ReadFile(login_data_file)
    if err!=nil{
        panic(err.Error())
    }

    // Turn file content into slice of structs
    var login_data []LoginData
    err=json.Unmarshal(content, &login_data)
    if err!=nil{
        panic(err.Error())
    }

    // Make login map (username -> all login data) from login data of all users
    login_map:=make(map[string]*LoginData)
    for i:=0; i<len(login_data); i++ {
        if _,ok:=login_map[login_data[i].Login_username];ok{
            panic("Duplicate username")
        }
        login_map[login_data[i].Login_username]=&login_data[i]
    }

    return LoginMap{login_map}
}

func (l *LoginMap) get_login_data(login_name string) (*LoginData, bool){
    login_data, ok:=l.login_map[login_name]
    return login_data, ok
}



type PlayerMap struct{
    player_map map[uint64]*PlayerData // (user_id -> his ingame player data)
    // player_data_modification_lock *sync.Mutex
}

// This handles the reading of user player data from disk
func load_PlayerMap() PlayerMap{
    // Read whole file
    content, err:=ioutil.ReadFile(player_data_file)
    if err!=nil{
        panic(err.Error())
    }

    // Turn file content into slice of structs
    var player_data []PlayerData
    err=json.Unmarshal(content, &player_data)
    if err!=nil{
        panic(err.Error())
    }

    // Make player map (id -> all player data) from player data of all users
    player_map:=make(map[uint64]*PlayerData)
    for i:=0; i<len(player_data); i++ {
        if _,ok:=player_map[player_data[i].Id];ok{
            panic("Duplicate player id")
        }
        player_map[player_data[i].Id]=&player_data[i]
    }

    return PlayerMap{player_map} // , &sync.Mutex{}
}

// func (p *PlayerMap) i_will_modify_player_data(){
//     p.player_data_modification_lock.Lock()
// }

// func (p *PlayerMap) i_am_done_modifying_player_data(){
//     p.player_data_modification_lock.Unlock()
// }
// maybe just a get function to get the mutex?

func (p *PlayerMap) get_player_data(id uint64) (*PlayerData, bool){
    player_data, ok:=p.player_map[id]
    return player_data, ok
}



func check_data_integrity(LM LoginMap, PM PlayerMap){
    login_map:=LM.login_map
    player_map:=PM.player_map

    // Make sure login_map and player_map had the same ids
    if len(login_map)!=len(player_map){
        panic("len(login_map)!=len(player_map)")
    } else {
        for _,login :=range login_map{
            if _,ok:=player_map[login.Id];!ok{
                panic("Id in login_map not present in player_map")
            }
        }
    }
}

type HandlerForFile struct{
    content []byte
    content_type string
}

func (h *HandlerForFile) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", h.content_type)
    w.Write(h.content)
}

// TODO: Log all errors in error log file
// TODO: Thread safety for all maps
//      actually, only maps that are modified on runtime
//      for player data: use locks only when writing, or read-write stuff. Dont use locks for only reads
func main() {
    login_map:=load_LoginMap()
    player_map:=load_PlayerMap()
    check_data_integrity(login_map, player_map)
    login_state_map:=new_LoginStateMap()
    world:=new_World()

    // Threadsafe rng
    rng:=make(chan int64)
    go func(){
        state:=[64]byte{}
        _,err:=rand.Read(state[:])
        if err!=nil{
            panic("Could not get random bytes")
        }

        now:=time.Now()
        state[0]^=byte(now.Unix())
        state[1]^=byte(now.UnixNano())

        s:=new_MySource(state)
        for{
            rng<-s.Int63()
        }
    }()

    get_ip:=func(r *http.Request)(string,error){
        i:=strings.LastIndex(r.RemoteAddr, ":")
        if i<0 {
            return "", errors.New("Weird address")
        }
        return r.RemoteAddr[:i], nil
    }

    get_player_data:=func(r *http.Request) (*PlayerData, error){
        // If ther is no cookie
        cookie,err:=r.Cookie(cookie_name)
        if err!=nil{
            return nil, errors.New("No cookie")
        }

        // If cookie cannot be split at /
        split_cookie:=strings.SplitN(cookie.Value, "/", 2)
        if len(split_cookie)!=2{
            return nil, errors.New("Ill formated cookie (split len)")
        }

        // If player id cannot be extracted from cookie
        player_id,err:=strconv.ParseUint(split_cookie[1], 16, 64)
        if err!=nil{
            return nil, errors.New("Ill formated cookie (parse)")
        }

        // If player id is not within the login_state_map
        login_state, ok:=login_state_map.get_login_state(player_id)
        if !ok{
            return nil, errors.New("Id not registered as logged in")
        }

        // If login token in cookie is different from the one in the saved login state
        if login_state.token!=cookie.Value{
            return nil, errors.New("Wrong login token")
        }

        // If the ip cannot be gotten from request
        ip,err:=get_ip(r)
        if err!=nil{
            return nil, err
        }

        // If the request IP is not the same as the one used to log in
        if ip!=login_state.ip{
            return nil, errors.New("Wrong ip")
        }

        // If the cookie should have expired
        if time.Now().After(login_state.expire_date){
            return nil, errors.New("Login expired")
        }

        return login_state.player_data, nil
    }

    mux := http.NewServeMux()

    // Makes all files in directory *available_files_dir* available for get request
    files, err:=ioutil.ReadDir(available_files_dir)
    if err!=nil{
        panic("Could not read directory of available files")
    }
    for _, file := range files {
        path:=fmt.Sprintf("%s/%s", available_files_dir, file.Name())
        content, err:=ioutil.ReadFile(path)
        if err!=nil{
            panic(err.Error())
        }

        content_type:=func() string{
            if strings.HasSuffix(file.Name(), ".html"){
                return "text/html"
            }else if strings.HasSuffix(file.Name(), ".jpg"){
                return "image/jpeg"
            }else if strings.HasSuffix(file.Name(), ".css"){
                return "text/css"
            }else if strings.HasSuffix(file.Name(), ".png"){
                return "image/png"
            }else if strings.HasSuffix(file.Name(), ".ico"){
                return "image/x-icon"
            } else{
                return ""
            }
        }()

        mux.Handle(fmt.Sprintf("/%s/%s", available_files_address, file.Name()), &HandlerForFile{content, content_type})
    }

    mux.HandleFunc(root_address, func (w http.ResponseWriter, r *http.Request){
        _,err:=get_player_data(r)
        if(err!=nil){
            http.Redirect(w, r, login_address, http.StatusFound)
            return
        }

        http.Redirect(w, r, home_address, http.StatusFound)
    })

    mux.HandleFunc(home_address, func (w http.ResponseWriter, r *http.Request){
        _,err:=get_player_data(r)
        if(err!=nil){
            http.Redirect(w, r, login_address, http.StatusFound)
            return
        }

        http.ServeFile(w, r, home_file)
    })

    // Logs the player in
    mux.HandleFunc(login_address, func (w http.ResponseWriter, r *http.Request){
        if r.Method=="GET"{
            http.ServeFile(w, r, login_file)
            return
        } else if r.Method=="POST" {
            username:=r.FormValue("username")
            password:=r.FormValue("password")

            // If form fields are not available or empty:
            if username=="" || password==""{
                http.Error(w, "Invalid login", http.StatusUnauthorized)
                return
            }

            // If user does not exist
            login_data,ok:=login_map.get_login_data(username)
            if !ok{
                http.Error(w, "Invalid login", http.StatusUnauthorized)
                return
            }

            is_password_correct:=func() bool{
                h := sha256.New()
                h.Write([]byte(password))
                h.Write([]byte(login_data.Salt))
                bs := h.Sum(nil)

                return fmt.Sprintf("%x", bs)==login_data.Salted_password_hash
            }()
            // If password is incorrect
            if !is_password_correct{
                http.Error(w, "Invalid login", http.StatusUnauthorized)
                return
            }

            //Get IP
            ip,err:=get_ip(r)
            if err!=nil{
                http.Error(w, "Weird address", http.StatusUnauthorized)
                return
            }

            expiration:=time.Now().Add(login_lifetime)
            cookie_value:=fmt.Sprintf("%x/%x", <-rng, login_data.Id)
            player_data,ok:=player_map.get_player_data(login_data.Id)
            if !ok{
                panic("Internal error (player_data,ok:player_map.get_player_data(login_data.Id); !ok)")
            }
            login_state:=LoginState{player_data, cookie_value, ip, expiration}
            login_state_map.set_login_state(login_data.Id, login_state)
            cookie:=http.Cookie{Name: cookie_name, Value: cookie_value, Expires: expiration}
            http.SetCookie(w, &cookie)
            http.Redirect(w, r, root_address, http.StatusFound)

            return
        } else {
            http.Error(w, "Request must ge GET or POST.", http.StatusBadRequest)
            return
        }
    })

    // Logs the player out
    mux.HandleFunc(unlogin_address, func (w http.ResponseWriter, r *http.Request){
        expiration:=time.Now().Add(-login_lifetime)
        cookie:=http.Cookie{Name: cookie_name, Value: "Value", Expires: expiration}
        http.SetCookie(w, &cookie)
        http.Redirect(w, r, root_address, http.StatusFound)
    })

    // mux.HandleFunc(get_name_address, func (w http.ResponseWriter, r *http.Request){
    //     ids, ok:=r.URL.Query()["id"]
    //     if !ok || len(ids)<1{
    //         http.Error(w, "Query error", http.StatusBadRequest)
    //         return
    //     }

    //     id,err:=strconv.ParseUint(ids[0], 10, 64)
    //     if err!=nil{
    //         http.Error(w, "Query error", http.StatusBadRequest)
    //         return
    //     }

    //     player_data, ok:=player_map[id]
    //     if !ok{
    //         http.Error(w, "Query error", http.StatusBadRequest)
    //         return
    //     }

    //     w.Write([]byte(player_data.Name))
    // })

    // Gives a player his player data
    mux.HandleFunc(get_data_address, func (w http.ResponseWriter, r *http.Request){
        player_data,err:=get_player_data(r)
        if(err!=nil){
            http.NotFound(w,r)
            return
        }

        json_player_data,err:=json.Marshal(player_data)
        if err!=nil{
            log.Println(err)
            http.Error(w, "Internal Error", http.StatusInternalServerError)
            return
        }

        w.Write(json_player_data)
    })

    // Gives a player a map of the fields near him
    mux.HandleFunc(get_world_address, func (w http.ResponseWriter, r *http.Request){
        player_data,err:=get_player_data(r)
        if(err!=nil){
            http.NotFound(w,r)
            return
        }

        var to_send struct{
            X int64
            Y int64
            World_array [21]byte
            Abundance_at_xy byte
        }

        x:=player_data.Pos_x
        y:=player_data.Pos_y
        to_send.X=x
        to_send.Y=y

        to_send.World_array[0],_=world.get_fields_static_data(x-1,y+2)
        to_send.World_array[1],_=world.get_fields_static_data(x,y+2)
        to_send.World_array[2],_=world.get_fields_static_data(x+1,y+2)

        to_send.World_array[3],_=world.get_fields_static_data(x-2,y+1)
        to_send.World_array[4],_=world.get_fields_static_data(x-1,y+1)
        to_send.World_array[5],_=world.get_fields_static_data(x,y+1)
        to_send.World_array[6],_=world.get_fields_static_data(x+1,y+1)
        to_send.World_array[7],_=world.get_fields_static_data(x+2,y+1)

        to_send.World_array[8],_=world.get_fields_static_data(x-2,y)
        to_send.World_array[9],_=world.get_fields_static_data(x-1,y)
        to_send.World_array[10],to_send.Abundance_at_xy=world.get_fields_static_data(x,y)
        to_send.World_array[11],_=world.get_fields_static_data(x+1,y)
        to_send.World_array[12],_=world.get_fields_static_data(x+2,y)

        to_send.World_array[13],_=world.get_fields_static_data(x-2,y-1)
        to_send.World_array[14],_=world.get_fields_static_data(x-1,y-1)
        to_send.World_array[15],_=world.get_fields_static_data(x,y-1)
        to_send.World_array[16],_=world.get_fields_static_data(x+1,y-1)
        to_send.World_array[17],_=world.get_fields_static_data(x+2,y-1)

        to_send.World_array[18],_=world.get_fields_static_data(x-1,y-2)
        to_send.World_array[19],_=world.get_fields_static_data(x,y-2)
        to_send.World_array[20],_=world.get_fields_static_data(x+1,y-2)

        json_player_data,err:=json.Marshal(to_send)
        if err!=nil{
            log.Println(err)
            http.Error(w, "Internal Error", http.StatusInternalServerError)
            return
        }

        w.Write(json_player_data)
    })

    if err:=http.ListenAndServe(":8000", mux);err!=nil{
        log.Println(err)
    }
}
