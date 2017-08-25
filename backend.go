package main;

import "log"
import "encoding/json"
import "net/http"
import "time"
import "fmt"
import "io/ioutil"
// import "crypto/sha256"
// import "crypto/sha512"
// import "crypto/md5"
import "strings"
import "errors"
import "strconv"
// import "sync"
import "crypto/rand"
import "os"

const(
    login_lifetime time.Duration = 2*time.Hour
    available_files_dir = "files"
    available_files_address = "files"
    world_seed_file = "data/world_seed"
    login_data_file = "data/logins.data"
    player_data_file = "data/players.data"
    login_file = "files/login.html"
    home_file = "elm_home/index.html"
    world_file = "elm_world/index.html"
    add_player_file = "files/add_player.html"
    modify_self_file = "files/modify_self.html"
    login_address = "/login"
    unlogin_address = "/unlogin"
    get_name_address = "/get_name"
    get_data_address = "/get_data"
    get_world_address = "/get_world"
    add_player_address = "/add_player"
    modify_self_address = "/modify_self"
    root_address = "/"
    home_address = "/home"
    world_address = "/world"
    cookie_name = "login_token"

    add_player_permission_flag = 0x0001
    modify_self_permission_flag = 0x0002

    // time_slot_size = 12*time.Hour
    steps_per_time_slot = 20
    // beginning_of_time = //something like time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
)


func check_data_integrity(LM LoginMap, PM PlayerMap){
    login_map:=LM.login_map
    player_map:=PM.player_map

    // Make sure login_map and player_map had the same ids
    if len(login_map)!=len(player_map){
        panic("len(login_map)!=len(player_map)")
    } else {
        for _,login :=range login_map{
            if _,ok:=player_map[login.id];!ok{
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

// TODO: log successful writes to savefiles
// TODO: under permissions also add infinite all (steps,resources,etc are not spent), transform data page (to set steps,resource,position,etc)
// TODO: add "check time, add steps if needed" function somewhere
// TODO: add counter to next time slot in home page and world page
// TODO: Log all errors in error log file
func main() {
    login_map:=load_LoginMap(login_data_file)
    player_map:=load_PlayerMap(player_data_file)
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

    get_player_data_and_permissions:=func(r *http.Request) (*PlayerData, uint16, error){
        // If ther is no cookie
        cookie,err:=r.Cookie(cookie_name)
        if err!=nil{
            return nil, 0, errors.New("No cookie")
        }

        // If cookie cannot be split at /
        split_cookie:=strings.SplitN(cookie.Value, "/", 2)
        if len(split_cookie)!=2{
            return nil, 0, errors.New("Ill formated cookie (split len)")
        }

        // If player id cannot be extracted from cookie
        player_id,err:=strconv.ParseUint(split_cookie[1], 16, 64)
        if err!=nil{
            return nil, 0, errors.New("Ill formated cookie (parse)")
        }

        // If player id is not within the login_state_map
        login_state, ok:=login_state_map.get_login_state(player_id)
        if !ok{
            return nil, 0, errors.New("Id not registered as logged in")
        }

        // If login token in cookie is different from the one in the saved login state
        if login_state.token!=cookie.Value{
            return nil, 0, errors.New("Wrong login token")
        }

        // If the ip cannot be gotten from request
        ip,err:=get_ip(r)
        if err!=nil{
            return nil, 0, err
        }

        // If the request IP is not the same as the one used to log in
        if ip!=login_state.ip{
            return nil, 0, errors.New("Wrong ip")
        }

        // If the cookie should have expired
        if time.Now().After(login_state.expire_date){
            return nil, 0, errors.New("Login expired")
        }

        return login_state.player_data, login_state.permissions, nil
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
        _,_,err:=get_player_data_and_permissions(r)
        if(err!=nil){
            http.Redirect(w, r, login_address, http.StatusFound)
            return
        }

        http.Redirect(w, r, home_address, http.StatusFound)
    })

    mux.HandleFunc(home_address, func (w http.ResponseWriter, r *http.Request){
        _,_,err:=get_player_data_and_permissions(r)
        if(err!=nil){
            http.Redirect(w, r, login_address, http.StatusFound)
            return
        }

        http.ServeFile(w, r, home_file)
    })

    mux.HandleFunc(world_address, func (w http.ResponseWriter, r *http.Request){
        _,_,err:=get_player_data_and_permissions(r)
        if(err!=nil){
            http.Redirect(w, r, login_address, http.StatusFound)
            return
        }

        http.ServeFile(w, r, world_file)
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

            // If password is incorrect
            if !login_data.is_password_correct(password){
                http.Error(w, "Invalid login", http.StatusUnauthorized)
                return
            }

            //Get IP
            ip,err:=get_ip(r)
            // If IP could not be gotten
            if err!=nil{
                http.Error(w, "Weird address", http.StatusUnauthorized)
                return
            }

            expiration:=time.Now().Add(login_lifetime)
            cookie_value:=fmt.Sprintf("%x/%x", <-rng, login_data.id)
            player_data,ok:=player_map.get_player_data(login_data.id)
            // If there is inconsistency in internal data structures *login_map* and *player_map*
            if !ok{
                panic("Internal error (player_data,ok:player_map.get_player_data(login_data.Id); !ok)")
            }
            login_state:=LoginState{player_data, login_data.permissions, cookie_value, ip, expiration}
            login_state_map.set_login_state(login_data.id, login_state)
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
        player_data,_,err:=get_player_data_and_permissions(r)
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
        player_data,_,err:=get_player_data_and_permissions(r)
        if(err!=nil){
            http.NotFound(w,r)
            return
        }

        // A post message to this address means the player wants to move
        if r.Method=="POST"{
            direction:=r.FormValue("direction")
            if direction=="Up"{
                player_map.move_up(player_data)
            } else if direction=="Down"{
                player_map.move_down(player_data)
            } else if direction=="Left"{
                player_map.move_left(player_data)
            } else if direction=="Right"{
                player_map.move_right(player_data)
            }
        }

        var to_send struct{
            X int64
            Y int64
            World_array [21]byte
            Abundance_at_xy byte
            Available_steps int64
        }

        x:=player_data.Pos_x
        y:=player_data.Pos_y
        to_send.X=x
        to_send.Y=y
        to_send.Available_steps=player_data.Available_steps

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

    mux.HandleFunc(add_player_address, func (w http.ResponseWriter, r *http.Request){
        _,permissions,err:=get_player_data_and_permissions(r)
        if(err!=nil){
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }

        if (permissions&add_player_permission_flag)==0{
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }

        if r.Method=="GET"{
            http.ServeFile(w, r, add_player_file)
            return
        } else if r.Method=="POST" {
            login_name:=r.FormValue("login_name")
            password:=r.FormValue("password")
            name:=r.FormValue("name")
            pos_x,e1:=strconv.ParseInt(r.FormValue("pos_x"), 10, 64)
            pos_y,e2:=strconv.ParseInt(r.FormValue("pos_y"), 10, 64)
            resource_a,e3:=strconv.ParseInt(r.FormValue("resource_a"), 10, 64)
            resource_b,e4:=strconv.ParseInt(r.FormValue("resource_b"), 10, 64)
            resource_c,e5:=strconv.ParseInt(r.FormValue("resource_c"), 10, 64)

            // If form fields are not available or empty or and error ocurred:
            if login_name=="" || password=="" || e1!=nil || e2!=nil || e3!=nil || e4!=nil || e5!=nil || len(login_name)>24 || len(name)>24{
                http.Error(w, "Input error", http.StatusBadRequest)
                return
            }
            // Add user
            id,err:=login_map.add_login(0, login_name, password, <-rng)
            if err!=nil{
                http.Error(w, fmt.Sprintf("Could not add user (%s)", err.Error()), http.StatusInternalServerError)
                return
            }
            // Add player
            err=player_map.add_player(id, name, pos_x, pos_y, resource_a, resource_b, resource_c)
            if err!=nil{
                http.Error(w, fmt.Sprintf("Could not add player (data is inconsistent now, %s)", err.Error()) , http.StatusInternalServerError)
                return
            }
            w.Write([]byte("Player added"))
            return
        } else {
            http.Error(w, "Request must ge GET or POST.", http.StatusBadRequest)
            return
        }
    })

    mux.HandleFunc(modify_self_address, func (w http.ResponseWriter, r *http.Request){
        player_data,permissions,err:=get_player_data_and_permissions(r)
        if(err!=nil){
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }

        if (permissions&modify_self_permission_flag)==0{
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }

        if r.Method=="GET"{
            http.ServeFile(w, r, modify_self_file)
            return
        } else if r.Method=="POST" {
            pos_x:=r.FormValue("pos_x")
            if pos_x!=""{
                player_data.Pos_x,_=strconv.ParseInt(pos_x,10,64)
            }

            pos_y:=r.FormValue("pos_y")
            if pos_y!=""{
                player_data.Pos_y,_=strconv.ParseInt(pos_y,10,64)
            }

            resource_a:=r.FormValue("resource_a")
            if resource_a!=""{
                player_data.Resource_A,_=strconv.ParseInt(resource_a,10,64)
            }

            resource_b:=r.FormValue("resource_b")
            if resource_b!=""{
                player_data.Resource_B,_=strconv.ParseInt(resource_b,10,64)
            }

            resource_c:=r.FormValue("resource_c")
            if resource_c!=""{
                player_data.Resource_C,_=strconv.ParseInt(resource_c,10,64)
            }

            available_steps:=r.FormValue("available_steps")
            if available_steps!=""{
                player_data.Available_steps,_=strconv.ParseInt(available_steps,10,64)
            }

            player_map.modify_player_file_chan<-func(f *os.File){
                if pos_x!=""{
                    save_pos_x(f, player_data)
                }
                if pos_y!=""{
                    save_pos_y(f, player_data)
                }

                if resource_a!=""{
                    save_resource_a(f, player_data)
                }

                if resource_b!=""{
                    save_resource_b(f, player_data)
                }

                if resource_c!=""{
                    save_resource_c(f, player_data)
                }

                if available_steps!=""{
                    save_available_steps(f, player_data)
                }
            }

            w.Write([]byte("Self modified"))
            return
        } else {
            http.Error(w, "Request must ge GET or POST.", http.StatusBadRequest)
            return
        }
    })

    log.Println("Start server")
    if err:=http.ListenAndServe(":8000", mux);err!=nil{
        log.Println(err)
    }
}
