package main;

import "log"
import "encoding/json"
import "net/http"
import "time"
import "fmt"
import "io/ioutil"
import "crypto/sha256"
import "crypto/sha512"

const(
    login_lifetime time.Duration = 2*time.Hour
    login_data_file = "data/login_data.json"
    login_file_path = "files/login.html"
    login_address = "/login"
    unlogin_address = "/unlogin"
    root_address = "/"
    cookie_name = "token"
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

type UserLoginData struct{
    Id uint64
    Name string
    Salted_password_hash string
    Salt string
}

type UserData struct{
    Id uint64
    Name string
    Pos_x int64
    Pos_y int64

}

func main() {
    login_map:=func() (map[string]*UserLoginData){
        content, err:=ioutil.ReadFile(login_data_file)
        if err!=nil{
            panic(err.Error())
        }

        var user_login_data []UserLoginData
        err=json.Unmarshal(content, &user_login_data)
        if err!=nil{
            panic(err.Error())
        }

        login_map:=make(map[string]*UserLoginData)
        for i:=0; i<len(user_login_data); i++ {
            login_map[user_login_data[i].Name]=&user_login_data[i]
        }

        return login_map
    }()




    rng:=make(chan int64)
    go func(){
        state:=[64]byte{99,78,221,9,123,79,7,50,9,3,5,37,91,3,2,8,69,21,64,123,62,31,45,59,124,32,98,23,74,56,5,3}
        s:=new_MySource(state)
        for{
            rng<-s.Int63()
        }
    }()




    mux := http.NewServeMux()
    mux.HandleFunc(root_address, func (w http.ResponseWriter, r *http.Request){
        cookie,err:=r.Cookie(cookie_name)
        if(err!=nil){
            http.Redirect(w, r, login_address, http.StatusFound)
            return
        }

        http.Error(w, cookie.Value, 403)
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

            expiration:=time.Now().Add(login_lifetime)
            cookie_value:=fmt.Sprintf("%s/%d", fmt.Sprintf("%x", <-rng), login_data.Id)
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

    if err:=http.ListenAndServe(":8000", mux);err!=nil{
        log.Println(err)
    }
}
