package main

import (
    "fmt"
    "net"
    "os"
    "bufio"
    "io"
    "time"
    "flag"
)

/* Version 1.0 Stable */

/* prints the prompt when or as requested */
func printPrompt(pChan chan uint32) {
    for run := true;run == true; {
        select {
        case <- pChan:
            fmt.Print("\n> ")
        default:
        }
        time.Sleep(20 * time.Millisecond)
    }
}

/* reads user input and handles it */
func readIn(conn net.Conn, runChan chan bool, pChan chan uint32) {
    reader := bufio.NewReader(os.Stdin)
    for {
        words, _, _ := reader.ReadLine()
        _, err := conn.Write(words)
        check(err)
        if string(words) == "!quit" {
            runChan <- false
            break
        }
        //pChan <- uint32(1)
        time.Sleep(20 * time.Millisecond)
    }
}

/* reads things from the server and handles them */
func readServer(conn net.Conn, runChan chan bool, pChan chan uint32) {
    //var ticker uint32 = 0
    for {
        words := make([]byte, 512)
        _, err := conn.Read(words)
        if err == io.EOF {
            fmt.Print("\nDisconnected from server.\n")
            runChan <- false
            break
        } else {
            check(err)
        }

        blank := []byte{0,0,0,0,0,0}
        cnt := 0
        for p, w := range blank {
            if words[p] == w {
                cnt++
            }
        }
        if cnt < 4 {
            fmt.Print(string(words), "\n")
            pChan <- uint32(1)
            time.Sleep(20 * time.Millisecond)
        }
    }
}

/* dialServer will connect to a pre-selected server */
func dialServer(target string, masterChan chan uint32) {
    runChan := make(chan bool, 1)
    pChan := make(chan uint32, 1)
    //var words []byte
    conn, err := net.Dial("tcp", target)
    if err != nil {
        fmt.Print(err, "\n")
    }
    /* get our info back from the server */
    go printPrompt(pChan)
    go readServer(conn, runChan, pChan)
    go readIn(conn, runChan, pChan)

    for run := true; run == true; {
        select {
        case <- runChan:
            close(runChan)
            close(pChan)
            masterChan <- 1
            run = false
        default:
        }
        time.Sleep(20 * time.Millisecond)
    }
}

/* check checks the error err for an error and crashes the program if != nil */
func check(err error) {
    if err != nil {
        panic(err)
    }
}

/* Simple raw connection client */

func main() {
    words := ""
    masterChan := make(chan uint32, 1)
    flag.StringVar(&words, "a", "localhost:9090", "Set address to dial to for server.")
    flag.Parse()
    defer fmt.Print("\nGoodbye!\n")

    fmt.Print("Dialing ", words, ";\n")
    fmt.Print("\n> ")
    go dialServer(words, masterChan)
    <- masterChan
}
