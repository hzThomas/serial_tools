package main

import (
    "bufio"
    "flag"
    "fmt"
    "github.com/tarm/goserial"
    "io"
    "log"
    "os"
    "strconv"
    "strings"
)

func recvUartData(s io.ReadWriteCloser) {
    for {
        buf := make([]byte, 128)
        n, err := s.Read(buf)
        if err != nil {
            log.Fatal(err)
        }
        fmt.Printf("%s", buf[:n])
    }
}

func openUart(name string, baud int) (io.ReadWriteCloser, error) {
    c := &serial.Config{Name: name, Baud: baud}

    s, err := serial.OpenPort(c)
    if err != nil {
        log.Fatal(err)
    }
    return s, nil
}

func closeUart(s io.ReadWriteCloser) {
    fmt.Println("Uart Closed")
    s.Close()
}

func sendUartData(s io.ReadWriteCloser, data []byte) {
    _, err := s.Write(data)
    if err != nil {
        log.Fatal(err)
    }
}

func sendUartString(s io.ReadWriteCloser, data string) {
    sendUartData(s, []byte(data+"\r\n"))
}

func sendUartHexData(s io.ReadWriteCloser, data string) {

    arr := strings.Split(strings.TrimSpace(data), " ")

    sl := make([]byte, len(arr))

    for i, value := range arr {

        if len(value) != 2 {
            fmt.Printf("Data Len ERROR[%s]\n", value)
            return
        }

        a, err := strconv.ParseInt(value, 16, 0)
        if err != nil {
            fmt.Printf("Data Context ERROR[%s]\n", value)
            return
        }
        sl[i] = byte(a)
    }
    sendUartData(s, sl)
}

func getInputCmd(reader *bufio.Reader) (string, string) {
    data, _, _ := reader.ReadLine()
    if len(data) == 0 {
        return "", ""
    }
    // 命令标识符
    if data[0] == ':' {
        cmdlist := strings.SplitN(string(data), " ", 2)
        if len(cmdlist) == 1 {
            return strings.ToUpper(cmdlist[0]), ""
        } else {
            return strings.ToUpper(cmdlist[0]), cmdlist[1]
        }
    } else {
        // 标识为直接输出数据，去除输入字符串前面多余的空格
        return ":O", strings.TrimSpace(string(data))
    }
}

func main() {
    uartName := flag.String("uart", "COM4", "set will use uart")
    uartBaud := flag.Int("baud", 115200, "set uart baudrate")
    flag.Parse()

    fmt.Printf("Open uart %s %d\n", *uartName, *uartBaud)
    s, err := openUart(*uartName, *uartBaud)
    if err != nil {
        log.Fatal(err)
    }
    defer closeUart(s)

    // 开启串口接收进程
    go recvUartData(s)

    running := true
    reader := bufio.NewReader(os.Stdin)
    for running {
        cmd, cmdParam := getInputCmd(reader)
        switch cmd {
        case ":Q":
            fallthrough
        case ":QUIT":
            running = false
        case ":":
            continue
        case "":
            continue
        case ":NOECHO":
            sendUartString(s, "ate0")
        case ":O":
            sendUartString(s, cmdParam)
        case ":H":
            sendUartHexData(s, cmdParam)
        default:
            fmt.Println("Error CMD")
        }
    }
}
