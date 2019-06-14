package main

import "fmt"
import "sync"
import "net"
import "bytes"
import "time"
import "flag"
import "math"

// config
var concurrency int = 100
var timeout int = 1
var verbose bool = true


func status(name string, done int, total int, timePassed int) {
    fmt.Printf("%s\t [%ds  %d/%d]\r", name, timePassed, done, total)
}

func generator(target []byte, alpha []byte, length int, ch chan []byte) {
    count := 0
    total := int(math.Pow(float64(len(alpha)), float64(length)))
    startTime := time.Now().Unix()
    data := make([]int, length)
    for {
        pos := length - 1
        result := make([]byte, length + len(target) + 1)
        for i, c := range data {
            result[i] = alpha[c]
        }
        result[len(data)] = '.'
        for i, c := range target {
            result[len(data) + 1 + i] = c
        }
        count += 1
        if count % 0x10 == 0 && verbose {
            timePassed := time.Now().Unix() - startTime

            go status(string(result), count, total, int(timePassed))
        }
        ch <- []byte(result)
        data[pos] += 1
        for data[pos] == len(alpha) {
            if pos == 0 {
                close(ch)
                return
            }
            data[pos] = 0
            data[pos - 1] += 1
            pos -= 1
        }
    }
}

func query(name []byte, dns string) bool {
    conn, err := net.Dial("udp", dns)
    if err != nil {
        return false
    }
    defer conn.Close()

    buffer := new(bytes.Buffer)

    // ID
    buffer.WriteByte('\x00')
    buffer.WriteByte('\x00')

    // FLAGS
    buffer.WriteByte('\x01')
    buffer.WriteByte('\x00')

    // QDCount, almost server never support multiple query in one record
    buffer.WriteByte('\x00')
    buffer.WriteByte('\x01')
    //buffer.WriteByte(byte(len(questions)))

    // ANCount
    buffer.WriteByte('\x00')
    buffer.WriteByte('\x00')

    // NSCount
    buffer.WriteByte('\x00')
    buffer.WriteByte('\x00')

    // ARCount
    buffer.WriteByte('\x00')
    buffer.WriteByte('\x00')

    // QName 
    for _, n := range bytes.Split(name, []byte{'.'}) {
        buffer.WriteByte(byte(len(n)))
        buffer.Write(n)
    }
    buffer.WriteByte('\x00')

    // QType A
    buffer.WriteByte('\x00')
    buffer.WriteByte('\x01')

    // QClass IN
    buffer.WriteByte('\x00')
    buffer.WriteByte('\x01')

    conn.Write(buffer.Bytes())

    resp := make([]byte, 8)
    conn.SetReadDeadline(time.Now().Add(time.Duration(timeout) * time.Second))
    _, err = conn.Read(resp)

    if err != nil {
        return false
    }

    return resp[3] & 0xF == 0 && resp[7] != 0
}

func brute(target []byte, alphabet []byte, length int, dns string) {
    source := make(chan []byte)
    go generator(target, alphabet, length, source)

    wg := new(sync.WaitGroup)
    for i := 0; i < concurrency; i++ {
        wg.Add(1)
        go func () {
            for name := range source {
                if query(name, fmt.Sprintf("%s:53", dns)) {
                    fmt.Printf("\033[K%s\n", string(name))
                }
            }
            wg.Done()
        }()
    }
    wg.Wait()
}

func main() {
    var target, alphabet, dns string
    var length int

    flag.StringVar(&target, "t", "", "Target You Want To Bruteforce")
    flag.StringVar(&alphabet, "a", "abcdefghijklmnopqrstuvwxyz", "Brute Alphabet")
    flag.IntVar(&length, "l", 3, "Sub Domain Name Length")
    flag.StringVar(&dns, "dns", "8.8.8.8", "DNS Server")
    flag.Parse()

    if len(target) == 0 {
        flag.PrintDefaults()
        return
    }

    brute([]byte(target), []byte(alphabet), length, dns)
}
