package main

import (
    "fmt"
    "github.com/nsf/termbox-go"
    "flag"
    "unicode/utf8"
    "net"
    "io"
    "time"
//    "strings"
//sc  "strconv"
)
// zero byte: "\x00"

/* Version 1.0 Stable */

/* message buffer to print to screen and possibly scroll in the future */
var msgs []string = make([]string, 50)
/* text typed by the user */
var txt []byte = make([]byte, 1024)
var posTxt uint32 = 0

var pos, lpos int

func stripZeroes(in []byte)([]byte) {
    blank := []byte{0}
    var i int
    for i=len(in)-1;i >= 0; i-- {
        if in[i] != blank[0] {
            break
        }
    }
    out := make([]byte, i+1)
    for h:=0; h <= len(out)-1; h++ {
        out[h] = in[h]
    }
    return out
}

/* send a message to the server */
func sendMsg(conn *net.Conn, newmsg string) {
    addMsg(newmsg)
    _, err := (*conn).Write([]byte(newmsg))
    check(err)
}

/* add a message to the stack */
func addMsg(newmsg string) {
    shiftMsgs()
    msgs[0] = newmsg
    fixMsgs()
}

/* fixes overflowing messages (>78 msgs) */
func fixMsgs() {
    //should operate only on msgs[0]
    str := msgs[0]
    numRunes := utf8.RuneCountInString(str)
    if posRunes:=0;numRunes > 78 {
        for posRunes=0;len(str) > 0;posRunes++ {
            _, size := utf8.DecodeRuneInString(str)
            str = str[size:]
            if posRunes > 78 {
                addMsg(str)
                break
            }
        }
    }
}

/* shifts all messages down, discarding the remainder message for loading new msg */
func shiftMsgs() {
    tmpMsgs := msgs

    for i:=len(tmpMsgs)-1;i >= 1;i-- {
        tmpMsgs[i] = msgs[i-1]
    }
    copy(tmpMsgs, msgs)
    msgs[0] = " "
}

/* populates messages with blank messages */
func popMsgs() {
    for pos := range msgs {
        msgs[pos] = " "
    }
}

/* clears user buffer */
func clearTxt() {
    for pos := range txt {
        txt[pos] = byte(' ')
    }
}

/* prints to pos x, y */
func tbPrint(x, y int, fg, bg termbox.Attribute, msg string) {
    for _, c := range msg {
            termbox.SetCell(x, y, c, fg, bg)
            x++
    }
}

/* draws the screen */
func draw(w, h int) {
    for {
        defer termbox.Flush()

        //top bar
        tbPrint(0,0,termbox.ColorBlue, termbox.ColorBlack, "╔")
        for y := 0; y < 1; y++ {
            for x := 1; x < w-1; x++ {
                tbPrint(x, y, termbox.ColorBlue, termbox.ColorBlack, "═")
            }
        }
        tbPrint(w-1,0,termbox.ColorBlue, termbox.ColorBlack, "╗")

        //body
        // maybe a parseMessages() or fixMessages() function?
        top, bot := 0, h-6
        left, right := 1, w-1
        /* handle messages */
        i:=0
        for y := bot; y > top; y-- {
            tbPrint(1, y, termbox.ColorCyan, termbox.ColorBlack, msgs[i])
            leng := utf8.RuneCountInString(msgs[i])
            for x := left+leng; x < right; x++ {
                tbPrint(x, y, termbox.ColorCyan, termbox.ColorBlack, " ")
            }
            i++
        }
        /* end message handling */

        //print edges
        for y := 1; y < h - 5; y++ {
            tbPrint(0, y, termbox.ColorBlue, termbox.ColorBlack, "║")
            tbPrint(w-1, y, termbox.ColorBlue, termbox.ColorBlack, "║")
        }

        //bottom bar
        tbPrint(0,h-5,termbox.ColorBlue, termbox.ColorBlack, "╚")
        for y := h-5; y < h-4; y++ {
            for x := 1; x < w-1; x++ {
                tbPrint(x, y, termbox.ColorBlue, termbox.ColorBlack, "═")
            }
        }
        tbPrint(w-1, h-5, termbox.ColorBlue, termbox.ColorBlack, "╝")

        //user input zone !-! should make this a for loop
        pos=w-1
        lpos=0
        tbPrint(0, h-4, termbox.ColorWhite, termbox.ColorBlack, string(txt[lpos:pos]))
        lpos=pos
        pos+=w-1
        tbPrint(0, h-3, termbox.ColorWhite, termbox.ColorBlack, string(txt[lpos:pos]))
        lpos=pos
        pos+=w-1
        tbPrint(0, h-2, termbox.ColorWhite, termbox.ColorBlack, string(txt[lpos:pos]))
        lpos=pos
        pos+=w-1
        tbPrint(0, h-1, termbox.ColorWhite, termbox.ColorBlack, string(txt[lpos:pos]))
        termbox.Flush()
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
            addMsg("\nDisconnected from server.\n")
            runChan <- false
            break
        } else {
            check(err)
        }

        wordsNew := stripZeroes(words)
        addMsg(string(wordsNew))
        pChan <- uint32(1)
        time.Sleep(20 * time.Millisecond)
    }
}

/* dialServer will connect to a pre-selected server */
func dialServer(msgChan chan string, target string) {
    runChan := make(chan bool, 1)
    pChan := make(chan uint32, 1)

    //var words []byte
    conn, err := net.Dial("tcp", target)
    if err != nil {
        check(err)
    }
    /* get our info back from the server */
    go readServer(conn, runChan, pChan)
    //go readIn(conn, runChan, pChan)

    for run := true; run == true; {
        select {
        case <- runChan:
            close(runChan)
            close(pChan)
            run = false
        case w := <- msgChan:
            sendMsg(&conn, w)
        case <- pChan:

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

/* client for SocketH rewritten in termbox-go */

func main() {
    words := ""
    flag.StringVar(&words, "a", "localhost:9090", "Set address to dial to for server.")
    flag.Parse()
    defer fmt.Print("\nGoodbye!\n")
    defer termbox.Close()
    msgChan := make(chan string, 1)
    popMsgs()
    go dialServer(msgChan, words)

    err := termbox.Init()
    if err != nil {
        fmt.Println(err)
    }
    termbox.SetInputMode(termbox.InputAlt)
    w, h := termbox.Size()

    termbox.Clear(termbox.ColorBlack, termbox.ColorBlack)
    termbox.Flush()
    go draw(w,h)
    termbox.Flush()
    for run:=true;run==true; {
        switch ev := termbox.PollEvent(); ev.Type {
            case termbox.EventKey:
                key := string(ev.Ch)

                    if ev.Key == termbox.KeyCtrlQ {
                        run=false
                        termbox.Flush()
                    } else if ev.Key == termbox.KeyEnter {

                        w := (string(txt[0:posTxt]))
                        msgChan <- w
                        clearTxt()
                        posTxt = 0

                    } else if ev.Key == termbox.KeyBackspace2 || ev.Key == termbox.KeyBackspace {
                        txt[posTxt] = byte(' ')
                        if posTxt > 0 {
                            txt[posTxt-1] = byte(' ')
                            posTxt-=1
                        }

                    } else {
                        letr, _ := utf8.DecodeRuneInString(key)
                        txt[posTxt] = byte(letr)
                        posTxt++
                    }
        }
    }
}
