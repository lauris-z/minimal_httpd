package main

import (
	"io"
	"os"
	"net"
	"fmt"
	"time"
	"bufio"
	"strconv"
	"strings"
)

var HTML_EXTENSION = ".html"
var code200 = "HTTP/1.1 200 OK"
var code404 = "HTTP/1.1 404 NOT FOUND"
var contentTypeOctet = "Content-Type: application/octet-stream"
var contentTypeText = "Content-Type: text/html"
var contentDispositionAttachmentWithoutFilename = "Content-Disposition: attachment; filename="
var ROOT = ""
var LOGS = ""

func main() {
	port := os.Args[1]
	ROOT = os.Args[2]
	LOGS = os.Args[3]
	initLogs()
	createListenerDataGetHttp(":"+port)
}

func initLogs() {
	present:=exists(LOGS)
	if(!present){
		file, err := os.Create(LOGS)
		if(err==nil) {
			file.Close()
		} else {
			fmt.Println(err)
		}
	}
}

func exists(path string) bool {
    _, err := os.Stat(path)
    if err == nil { return true }
    if os.IsNotExist(err) { return false }
    return true
}

func storeLog(line string) {
	f, err := os.OpenFile(LOGS, os.O_APPEND|os.O_WRONLY, 0600)
  	if(err==nil) {
  		defer f.Close()
		timeInt64:=time.Now().UnixNano()/int64(time.Millisecond)
		timeString:=strconv.FormatInt(timeInt64, 10)
		f.WriteString(timeString+" "+line+"\n")
	} else {
		fmt.Println(err)
		if(err.Error()=="open "+LOGS+": no such file or directory") {
			initLogs()
			storeLog(line)
		}
	}
}

func createListenerDataGetHttp(port string) {
	listener, _ := net.Listen("tcp", port)
	for {
		connection, _ := listener.Accept()
		go doRequest(connection)
	}
}

func doRequest(connection net.Conn) {
	reader := bufio.NewReader(connection)
	line, _ := reader.ReadString('\n')
	words := strings.Fields(line)
	file:="index.html"
	if(len(words)>1) {
		file=words[1]
		if(file=="/") {
			file="/index.html"
		} else if(len(file)>1) {
			if(file[1:2]=="/") {
				file="/index.html" //fix bug if url = localhost:8080//toto.html (more than one slash)
			}
		}
		response:="ERROR"
		if(words[0]=="GET") {
			if _, err := os.Stat(ROOT+"/"+file[1:]); err != nil {
				response:=code404+"\n"+"\n"
				go storeLog(connection.RemoteAddr().String()+" GET "+file+" 404")
				connection.Write([]byte(response))
			} else {
				if(file[len(file)-len(HTML_EXTENSION):]==HTML_EXTENSION) {
					response=code200+"\n"+contentTypeText+"\n"+"\n"
					go storeLog(connection.RemoteAddr().String()+" GET "+file+" 200")
					connection.Write([]byte(response))
					filescan, err := os.Open(ROOT+"/"+file[1:])
					if err != nil {
						fmt.Println(err)
					}
					defer filescan.Close()
					scanner := bufio.NewScanner(filescan)
					line:=""
					for scanner.Scan() {
						line=scanner.Text()
						connection.Write([]byte(line))
					}
					filescan.Close()
				} else {
					contentDispositionAttachment:=contentDispositionAttachmentWithoutFilename+file[1:]
					openfile, _ := os.Open(ROOT+"/"+file[1:])
					data := make([]byte, 1048576)
					response=code200+"\n"+contentTypeOctet+"\n"+contentDispositionAttachment+"\n"+"\n"
					go storeLog(connection.RemoteAddr().String()+" GET "+file+" 200")
					connection.Write([]byte(response))
					k:=0
					total:=0
					for {
							count, err := openfile.Read(data)
							if err == io.EOF {
								break
							}
							k+=1
							if count < 1048576 {
								data2 := make([]byte, count)
								copy(data2,data)
								total+=len(data2)
								connection.Write([]byte(data2))
							} else {
								total+=len(data)
								connection.Write([]byte(data))
							}
					}
					openfile.Close()
				}
			}
		}
	}
	connection.Close()
}
