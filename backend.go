package main;

import "log"
import "encoding/json"
import "net/http"
import "time"
import "fmt"
import "io/ioutil"
import "crypto/sha256"
import "crypto/sha512"
import "strings"
import "errors"
import "strconv"

const(
    login_lifetime time.Duration = 2*time.Hour
    login_data_file = "data/login_data.json"
    player_data_file = "data/player_data.json"
    login_file_path = "files/login.html"
    login_address = "/login"
    unlogin_address = "/unlogin"
    get_name_address = "/get_name"
    get_data_address = "/get_data"
    root_address = "/"
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

// TODO: Log all errors in error log file
// TODO: Secret initial rng state should not be hardcoded
func main() {
    // This handles the reading of user login data from disk
    login_map:=func() map[string]*LoginData {
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

        return login_map
    }()

    // This handles the reading of user player data from disk
    player_map:=func() map[uint64]*PlayerData {
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

        return player_map
    }()

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

    // fmt.Println(login_map)
    // fmt.Println(player_map)

    login_state_map:=make(map[uint64]*LoginState)

    // Threadsafe rng
    rng:=make(chan int64)
    go func(){
        state:=[64]byte{99,78,221,9,123,79,7,50,9,3,5,37,91,3,2,8,69,21,64,123,62,31,45,59,124,32,98,23,74,56,5,3,
            2,4,8,3,7,6,0,1,2,5,5,2,5,3,6,2,3,3,3,0,9,7,0,0,9,7,7,2,3,3,7,3} // Secret initial state
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
        cookie,err:=r.Cookie(cookie_name)
        if err!=nil{
            return nil, errors.New("No cookie")
        }

        split_cookie:=strings.SplitN(cookie.Value, "/", 2)
        if len(split_cookie)!=2{
            return nil, errors.New("Ill formated cookie (split len)")
        }

        player_id,err:=strconv.ParseUint(split_cookie[1], 16, 64)
        if err!=nil{
            return nil, errors.New("Ill formated cookie (parse)")
        }

        login_state, ok:=login_state_map[player_id]
        if !ok{
            return nil, errors.New("Id not registered as logged in")
        }

        if login_state.token!=cookie.Value{
            return nil, errors.New("Wrong login token")
        }

        ip,err:=get_ip(r)
        if err!=nil{
            return nil, err
        }

        if ip!=login_state.ip{
            return nil, errors.New("Wrong ip")
        }

        if time.Now().After(login_state.expire_date){
            return nil, errors.New("Login expired")
        }

        return login_state.player_data, nil
    }

    mux := http.NewServeMux()
    mux.HandleFunc(root_address, func (w http.ResponseWriter, r *http.Request){
        player_data,err:=get_player_data(r)
        if(err!=nil){
            http.Redirect(w, r, login_address, http.StatusFound)
            return
        }

        w.Write([]byte(fmt.Sprintln(player_data)))

        return
    })

    mux.HandleFunc(login_address, func (w http.ResponseWriter, r *http.Request){
        if r.Method=="GET"{
            http.ServeFile(w, r, login_file_path)
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
            login_data,ok:=login_map[username]
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
            login_state_map[login_data.Id]=&LoginState{player_map[login_data.Id], cookie_value, ip, expiration}
            cookie:=http.Cookie{Name: cookie_name, Value: cookie_value, Expires: expiration}
            http.SetCookie(w, &cookie)
            http.Redirect(w, r, root_address, http.StatusFound)

            return
        } else {
            http.Error(w, "Request must ge GET or POST.", http.StatusBadRequest)
            return
        }
    })

    mux.HandleFunc(unlogin_address, func (w http.ResponseWriter, r *http.Request){
        expiration:=time.Now().Add(-login_lifetime)
        cookie:=http.Cookie{Name: cookie_name, Value: "Value", Expires: expiration}
        http.SetCookie(w, &cookie)
        http.Redirect(w, r, root_address, http.StatusFound)
    })

    mux.HandleFunc(get_name_address, func (w http.ResponseWriter, r *http.Request){
        ids, ok:=r.URL.Query()["id"]
        if !ok || len(ids)<1{
            http.Error(w, "Query error", http.StatusBadRequest)
            return
        }

        id,err:=strconv.ParseUint(ids[0], 10, 64)
        if err!=nil{
            http.Error(w, "Query error", http.StatusBadRequest)
            return
        }

        player_data, ok:=player_map[id]
        if !ok{
            http.Error(w, "Query error", http.StatusBadRequest)
            return
        }

        w.Write([]byte(player_data.Name))
    })

    mux.HandleFunc(get_data_address, func (w http.ResponseWriter, r *http.Request){
        player_data,err:=get_player_data(r)
        if(err!=nil){
            player_data=&PlayerData{}
        }

        json_player_data,err:=json.Marshal(player_data)
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
