package main;

// import "log"
// import "encoding/json"
// import "net/http"
// import "time"
// import "fmt"
// import "io/ioutil"
// import "crypto/sha256"
// import "crypto/sha512"
// import "crypto/md5"
// import "strings"
import "errors"
// import "strconv"
import "sync"
// import "crypto/rand"
import "os"


type PlayerData struct{
    file_pos int64
    read_and_modify_lock *sync.Mutex
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


type PlayerMap struct{
    player_map map[uint64]*PlayerData // (user_id -> his ingame player data)
    player_add_lock *sync.RWMutex
    add_player_file_chan chan func(*os.File, int64)
    modify_player_file_chan chan func(*os.File)
}

// This handles the reading of user player data from disk
func load_PlayerMap(filename string) PlayerMap{
    f, err:=os.Open(filename)
    if err!=nil{
        panic("Could not open login file")
    }
    defer f.Close()

    player_map:=make(map[uint64]*PlayerData)
    buffer:=make([]byte, 88)
    pos:=int64(0)
    for{
        n,err:=f.Read(buffer[:])
        if err!=nil{
            if n!=0{
                panic("Incomplete portion of data read from login file")
            }
            break
        }

        player_data:=&PlayerData{pos, &sync.Mutex{}, byte_slice_to_uint64(buffer[0:8]), byte_slice24_to_string(buffer[8:32]), byte_slice_to_int64(buffer[32:40]), byte_slice_to_int64(buffer[40:48]), byte_slice_to_int64(buffer[48:56]), byte_slice_to_int64(buffer[56:64]), byte_slice_to_int64(buffer[64:72]), byte_slice_to_int64(buffer[72:80]), byte_slice_to_int64(buffer[80:88])}
        pos+=88
        _,ok:=player_map[player_data.Id]
        if ok{
            panic("Duplicate id (load_PlayerMap)")
        }
        player_map[player_data.Id]=player_data
    }

    // A nice way to solve concurrent file access: only one thread does the accessing
    add_player_file_chan:=make(chan func(*os.File, int64), 20)
    modify_player_file_chan:=make(chan func(*os.File), 20)
    go func(){
        for{
            select{
            case add_player_func:=<-add_player_file_chan:
                f, err:=os.OpenFile(filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
                if err!=nil{
                    panic("Could not open file (add_player_file_chan)")
                }

                // Get stats of the opened file. Most importantly its size
                file_stat, err:=f.Stat()
                if err!=nil{
                    panic("Could not get file stats (add_player_file_chan)")
                }

                add_player_func(f, file_stat.Size())
                f.Close()

            case modify_player_func:=<-modify_player_file_chan:
                f, err:=os.OpenFile(filename, os.O_WRONLY, 0644)
                if err!=nil{
                    panic("Could not open file (modify_player_file_chan)")
                }

                modify_player_func(f)
                f.Close()
            }
        }
    }()

    return PlayerMap{player_map, &sync.RWMutex{}, add_player_file_chan, modify_player_file_chan}
}

func (p *PlayerMap) get_player_data(id uint64) (*PlayerData, bool){
    p.player_add_lock.RLock()
    player_data, ok:=p.player_map[id]
    p.player_add_lock.RUnlock()

    return player_data, ok
}


func save_pos_x(f*os.File, player_data *PlayerData){
    pos_x:=int64_to_bytes(player_data.Pos_x)
    _,err:=f.WriteAt(pos_x[:], player_data.file_pos+32)
    if err!=nil{
        panic("Could not write to file (save_pos_x)")
    }
}

func save_pos_y(f*os.File, player_data *PlayerData){
    pos_y:=int64_to_bytes(player_data.Pos_y)
    _,err:=f.WriteAt(pos_y[:], player_data.file_pos+40)
    if err!=nil{
        panic("Could not write to file (save_pos_y)")
    }
}

func save_resource_a(f*os.File, player_data *PlayerData){
    resource_a:=int64_to_bytes(player_data.Resource_A)
    _,err:=f.WriteAt(resource_a[:], player_data.file_pos+48)
    if err!=nil{
        panic("Could not write to file (save_resource_a)")
    }
}

func save_resource_b(f*os.File, player_data *PlayerData){
    resource_b:=int64_to_bytes(player_data.Resource_B)
    _,err:=f.WriteAt(resource_b[:], player_data.file_pos+56)
    if err!=nil{
        panic("Could not write to file (save_resource_b)")
    }
}

func save_resource_c(f*os.File, player_data *PlayerData){
    resource_c:=int64_to_bytes(player_data.Resource_C)
    _,err:=f.WriteAt(resource_c[:], player_data.file_pos+64)
    if err!=nil{
        panic("Could not write to file (save_resource_c)")
    }
}

func save_available_steps(f*os.File, player_data *PlayerData){
    available_steps:=int64_to_bytes(player_data.Available_steps)
    _,err:=f.WriteAt(available_steps[:], player_data.file_pos+72)
    if err!=nil{
        panic("Could not write to file (save_available_steps)")
    }
}

func save_last_time_gotten_steps(f*os.File, player_data *PlayerData){
    last_time_gotten_steps:=int64_to_bytes(player_data.Last_time_gotten_steps)
    _,err:=f.WriteAt(last_time_gotten_steps[:], player_data.file_pos+80)
    if err!=nil{
        panic("Could not write to file (save_last_time_gotten_steps)")
    }
}

func (p *PlayerMap) move_up(player_data *PlayerData) error{
    player_data.read_and_modify_lock.Lock()
    defer player_data.read_and_modify_lock.Unlock()

    if player_data.Available_steps<1{
        return errors.New("No available steps left")
    }

    player_data.Pos_y+=1
    player_data.Available_steps-=1

    p.modify_player_file_chan<-func(f *os.File){
        save_pos_y(f, player_data)
        save_available_steps(f, player_data)
    }

    return nil
}

func (p *PlayerMap) move_down(player_data *PlayerData) error{
    player_data.read_and_modify_lock.Lock()
    defer player_data.read_and_modify_lock.Unlock()

    if player_data.Available_steps<1{
        return errors.New("No available steps left")
    }

    player_data.Pos_y-=1
    player_data.Available_steps-=1

    p.modify_player_file_chan<-func(f *os.File){
        save_pos_y(f, player_data)
        save_available_steps(f, player_data)
    }

    return nil
}

func (p *PlayerMap) move_left(player_data *PlayerData) error{
    player_data.read_and_modify_lock.Lock()
    defer player_data.read_and_modify_lock.Unlock()

    if player_data.Available_steps<1{
        return errors.New("No available steps left")
    }

    player_data.Pos_x-=1
    player_data.Available_steps-=1

    p.modify_player_file_chan<-func(f *os.File){
        save_pos_x(f, player_data)
        save_available_steps(f, player_data)
    }

    return nil
}

func (p *PlayerMap) move_right(player_data *PlayerData) error{
    player_data.read_and_modify_lock.Lock()
    defer player_data.read_and_modify_lock.Unlock()

    if player_data.Available_steps<1{
        return errors.New("No available steps left")
    }

    player_data.Pos_x+=1
    player_data.Available_steps-=1

    p.modify_player_file_chan<-func(f *os.File){
        save_pos_x(f, player_data)
        save_available_steps(f, player_data)
    }

    return nil
}

func (p *PlayerMap) add_player(id uint64, name string, pos_x, pos_y, resource_a, resource_b, resource_c int64) error{
    if len(name)>24{
        return errors.New("Name is too long (len(name)>24, add_player)")
    }

    // Create new player
    available_steps:=int64(steps_per_time_slot)
    last_time_gotten_steps:=int64(0)
    new_player_data:=&PlayerData{-1000, &sync.Mutex{}, id, name, pos_x, pos_y, resource_a, resource_b, resource_c, available_steps, last_time_gotten_steps}

    p.player_add_lock.Lock()
    defer p.player_add_lock.Unlock()

    _,ok:=p.player_map[id] // Check whether id already exists in map
    if ok{
        return errors.New("Id already exists (add_player)")
    }

    // Add new player to player_map
    p.player_map[id]=new_player_data

    // Save new player to file (or rather, put the *saving* in a queue)
    p.add_player_file_chan<-func(f *os.File, file_pos int64){
        new_player_data.file_pos=file_pos
        f.Write(uint64_to_bytes(id))
        f.Write(string_to_24_bytes(name))
        f.Write(int64_to_bytes(pos_x))
        f.Write(int64_to_bytes(pos_y))
        f.Write(int64_to_bytes(resource_a))
        f.Write(int64_to_bytes(resource_b))
        f.Write(int64_to_bytes(resource_c))
        f.Write(int64_to_bytes(available_steps))
        f.Write(int64_to_bytes(last_time_gotten_steps))
    }

    return nil
}
