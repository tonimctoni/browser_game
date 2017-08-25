package main;


func uint16_to_bytes(n uint16) []byte {
    return []byte{byte(n>>8),byte(n>>0)}
}

func uint64_to_bytes(n uint64) []byte {
    return []byte{byte(n>>56),byte(n>>48),byte(n>>40),byte(n>>32),byte(n>>24),byte(n>>16),byte(n>>8),byte(n>>0)}
}

func int64_to_bytes(n int64) []byte {
    return []byte{byte(n>>56),byte(n>>48),byte(n>>40),byte(n>>32),byte(n>>24),byte(n>>16),byte(n>>8),byte(n>>0)}
}

func string_to_24_bytes(s string) []byte{
    if len(s)>24{
        panic("string length > 24")
    }

    ret:=make([]byte, 24)
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