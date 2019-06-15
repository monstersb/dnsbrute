package main

import "fmt"
import "sync"
import "net"
import "bytes"
import "time"
import "flag"
import "math"
import "os"
import "bufio"
import "io"

// config
var concurrency int = 200
var timeout int = 1
var dns string = "8.8.8.8:53"

var error int = 0
var success int = 0
var cost int = 0


func status(name string, done int, total int) {
    fmt.Printf("%s\t [%.2f%%] [cost: %ds, success: %d, error: %d]\r", name, float32(done) / float32(total) * 100, cost, success, error)
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
        if count % 0x30 == 0 {
            cost = int(time.Now().Unix() - startTime)
            go status(string(result), count, total)
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

func query(name []byte) bool {
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
        error += 1
        return false
    }

    return resp[3] & 0xF == 0 && resp[7] != 0
}

func brute(target []byte, alphabet []byte, length int, file io.Writer) {
    source := make(chan []byte)
    go generator(target, alphabet, length, source)

    wg := new(sync.WaitGroup)
    for i := 0; i < concurrency; i++ {
        wg.Add(1)
        go func () {
            for name := range source {
                if query(name) {
                    fmt.Fprintln(file, string(name))
                    success += 1
                }
            }
            wg.Done()
        }()
    }
    wg.Wait()
}

func main() {
    var target, alphabet, outFile string
    var length int

    flag.StringVar(&target, "t", "", "Target You Want To Bruteforce")
    flag.StringVar(&alphabet, "a", "abcdefghijklmnopqrstuvwxyz", "Brute Alphabet")
    flag.IntVar(&length, "l", 3, "Sub Domain Name Length")
    flag.StringVar(&outFile, "o", "output.txt", "Output File")
    flag.Parse()

    if len(target) == 0 {
        flag.PrintDefaults()
        return
    }

    file, err := os.Create(outFile)
    if err != nil {
        fmt.Println("A error occured while open", outFile)
        return
    }
    defer file.Close()
    fileBufer := bufio.NewWriter(file)

    brute([]byte(target), []byte(alphabet), length, fileBufer)
    fileBufer.Flush()

    fmt.Printf("\033[K")
    fmt.Println("Total:  ", math.Pow(float64(len(alphabet)), float64(length)))
    fmt.Println("Result: ", success)
    fmt.Println("Error:  ", success, "timeout")
    fmt.Println("cost:   ", cost, "seconds")
}
