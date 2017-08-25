package main;

import "os"
import "crypto/sha256"
import "crypto/rand"


const(
    login_file = "logins.data"
    player_file = "players.data"
)


type Entry struct{
    Id uint64
    Permissions uint16
    Login_username string
    Password string
    Name string
    Pos_x int64
    Pos_y int64
    Resource_A int64
    Resource_B int64
    Resource_C int64
    Available_steps int64
    Last_time_gotten_steps int64
}

func uint16_to_bytes(n uint16) [2]byte {
    return [2]byte{byte(n>>8),byte(n>>0)}
}

func uint64_to_bytes(n uint64) [8]byte {
    return [8]byte{byte(n>>56),byte(n>>48),byte(n>>40),byte(n>>32),byte(n>>24),byte(n>>16),byte(n>>8),byte(n>>0)}
}

func int64_to_bytes(n int64) [8]byte {
    return [8]byte{byte(n>>56),byte(n>>48),byte(n>>40),byte(n>>32),byte(n>>24),byte(n>>16),byte(n>>8),byte(n>>0)}
}

func string_to_24_bytes(s string) [24]byte{
    if len(s)>24{
        panic("string length > 24")
    }

    ret:=[24]byte{}
    for i:=0; i<24;i ++{
        if i<len(s){
            ret[i]=([]byte(s))[i]
        } else{
            ret[i]=0
        }
        
    }

    return ret
}

func byte_slice_to_uint16(b []byte) uint16{
    if len(b)!=2{
        panic("Byte slice's length != 2")
    }

    return uint16(b[0])<<8 | uint16(b[1])<<0
}

func byte_slice_to_uint64(b []byte) uint64{
    if len(b)!=8{
        panic("Byte slice's length != 8")
    }

    return uint64(b[0])<<56 | uint64(b[1])<<48 | uint64(b[2])<<40 | uint64(b[3])<<32 |
        uint64(b[4])<<24 | uint64(b[5])<<16 | uint64(b[6])<<8 | uint64(b[7])<<0
}

func byte_slice_to_int64(b []byte) int64{
    if len(b)!=8{
        panic("Byte slice's length != 8")
    }

    return int64(b[0])<<56 | int64(b[1])<<48 | int64(b[2])<<40 | int64(b[3])<<32 |
        int64(b[4])<<24 | int64(b[5])<<16 | int64(b[6])<<8 | int64(b[7])<<0
}

func byte_slice24_to_string(b []byte) string{
    if len(b)!=24{
        panic("Byte slice's length != 24")
    }
    l:=0
    for _,e:=range b{
        if e==0{
            break
        }
        l+=1
    }

    return string(b[:l])
}

func byte_slice32_to_array32(b []byte) [32]byte{
    if len(b)!=32{
        panic("Byte slice's length != 32")
    }

    ret:=[32]byte{}
    for i:=0; i<32; i++{
        ret[i]=b[i]
    }

    return ret
}

func byte_slice8_to_array8(b []byte) [8]byte{
    if len(b)!=8{
        panic("Byte slice's length != 8")
    }

    ret:=[8]byte{}
    for i:=0; i<8; i++{
        ret[i]=b[i]
    }

    return ret
}

func add_entry(entry Entry) {
    func(){
        f, err:=os.OpenFile(login_file, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
        if err!=nil{
            panic("Could not open file")
        }
        defer f.Close()

        id:=uint64_to_bytes(entry.Id)
        f.Write(id[:])

        permissions:=uint16_to_bytes(entry.Permissions)
        f.Write(permissions[:])

        // name_hash:=sha256.Sum256([]byte(entry.Login_username))
        name:=string_to_24_bytes(entry.Login_username)
        f.Write(name[:])

        salt:=[8]byte{}
        _,err=rand.Read(salt[:])
        if err!=nil{
            panic("Could not read random data")
        }
        f.Write(salt[:])

        h := sha256.New()
        h.Write([]byte(entry.Password))
        h.Write(salt[:])
        salted_password_hash := h.Sum(nil)
        f.Write(salted_password_hash[:])
    }()

    func(){
        f, err:=os.OpenFile(player_file, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
        if err!=nil{
            panic("Could not open file (2)")
        }
        defer f.Close()

        id:=uint64_to_bytes(entry.Id)
        f.Write(id[:])

        name:=string_to_24_bytes(entry.Name)
        f.Write(name[:])

        pos_x:=int64_to_bytes(entry.Pos_x)
        f.Write(pos_x[:])

        pos_y:=int64_to_bytes(entry.Pos_y)
        f.Write(pos_y[:])

        resource_a:=int64_to_bytes(entry.Resource_A)
        f.Write(resource_a[:])

        resource_b:=int64_to_bytes(entry.Resource_B)
        f.Write(resource_b[:])

        resource_c:=int64_to_bytes(entry.Resource_C)
        f.Write(resource_c[:])

        available_steps:=int64_to_bytes(entry.Available_steps)
        f.Write(available_steps[:])

        last_time_gotten_steps:=int64_to_bytes(entry.Last_time_gotten_steps)
        f.Write(last_time_gotten_steps[:])
    }()
}


func main() {
    add_entry(Entry{1, 3, "toni", "password", "the toniman", 0, 0, 100, 150, 25, 20, 0})
    add_entry(Entry{2, 0, "sabrina", "password", "sorceress_of_binaheim", 2, -2, 120, 75, 50, 18, 0})
}

// type LoginData struct{
//     id uint64
//     permissions uint16
//     login_username string
//     salt [8]byte
//     salted_password_hash [32]byte
// }

// // Just for some encapsulation <== program is getting complex
// type LoginMap struct{
//     login_map map[string]*LoginData // (username -> all login data)
// }

// // This handles the reading of user login data from disk
// func load_LoginMap() LoginMap{
//     f, err:=os.Open(login_file)
//     if err!=nil{
//         panic("Could not open login file")
//     }
//     defer f.Close()

//     login_map:=make(map[string]*LoginData)
//     buffer:=make([]byte, 74)
//     for{
//         n,err:=f.Read(buffer[:])
//         if err!=nil{
//             if n!=0{
//                 panic("Incomplete portion of data read from login file")
//             }
//             break
//         }

//         login_data:=&LoginData{byte_slice_to_uint64(buffer[0:8]),byte_slice_to_uint16(buffer[8:10]),byte_slice24_to_string(buffer[10:34]),byte_slice8_to_array8(buffer[34:42]),byte_slice32_to_array32(buffer[42:74])}

//         login_map[login_data.login_username]=login_data
//     }

//     return LoginMap{login_map}
// }

// type PlayerData struct{
//     id uint64
//     name string
//     pos_x int64
//     pos_y int64
//     resource_A int64
//     resource_B int64
//     resource_C int64
//     available_steps int64
//     last_time_gotten_steps int64 // number of time slots elapsed since beginning
// }

// type PlayerMap struct{
//     player_map map[uint64]*PlayerData // (user_id -> his ingame player data)
//     // player_data_modification_lock *sync.Mutex
// }

// // This handles the reading of user player data from disk
// func load_PlayerMap() PlayerMap{
//     f, err:=os.Open(player_file)
//     if err!=nil{
//         panic("Could not open login file")
//     }
//     defer f.Close()

//     player_map:=make(map[uint64]*PlayerData)
//     buffer:=make([]byte, 88)
//     for{
//         n,err:=f.Read(buffer[:])
//         if err!=nil{
//             if n!=0{
//                 panic("Incomplete portion of data read from login file")
//             }
//             break
//         }

//         player_data:=&PlayerData{byte_slice_to_uint64(buffer[0:8]), byte_slice24_to_string(buffer[8:32]), byte_slice_to_int64(buffer[32:40]), byte_slice_to_int64(buffer[40:48]), byte_slice_to_int64(buffer[48:56]), byte_slice_to_int64(buffer[56:64]), byte_slice_to_int64(buffer[64:72]), byte_slice_to_int64(buffer[72:80]), byte_slice_to_int64(buffer[80:88])}

//         player_map[player_data.id]=player_data
//     }

//     return PlayerMap{player_map}
// }

// func main() {
//     load_LoginMap()
//     load_PlayerMap()
// }