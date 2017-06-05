package main

import (
	"os"
	"net"
	"fmt"
	"io"
	"bufio"
	"encoding/binary"
	"strings"
	"time"
	"strconv"
)

var HTML_EXTENSION = ".html"
var LOGS = "logs_minimal_httpd.txt"

func main() {
	initLogs()
	port := os.Args[1]
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
  for {
		reader := bufio.NewReader(connection)
    line, _ := reader.ReadString('\n')
    words := strings.Fields(line)
		file:="robots.txt"
    if(len(words)>1) {
    	file=words[1]
    }
		response:="ERROR"
		code200:="HTTP/1.1 200 OK"
		code404:="HTTP/1.1 404 NOT FOUND"
		contentTypeOctet:="Content-Type: application/octet-stream"
		contentTypeText:="Content-Type: text/html"
		contentDispositionAttachment:="Content-Disposition: attachment; filename="+file[1:]
		if(len(words)>0) {
			if(words[0]=="GET") {
		    if _, err := os.Stat(file[1:]); err != nil {
		      response:=code404+"\n"+"\n"
		      go storeLog(connection.RemoteAddr().String()+" GET "+file+" 404")
		      connection.Write([]byte(response))
		      connection.Close()
		    } else {
					//fmt.Println("test="+file[len(file)-len(HTML_EXTENSION):])
					if(file[len(file)-len(HTML_EXTENSION):]==HTML_EXTENSION) {
		      	response=code200+"\n"+contentTypeText+"\n"+"\n"
						go storeLog(connection.RemoteAddr().String()+" GET "+file+" 200")
						connection.Write([]byte(response))
						filescan, err := os.Open(file[1:])
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
					} else {
						openfile, _ := os.Open(file[1:])
						data := make([]byte, 1048576)
						response=code200+"\n"+contentTypeOctet+"\n"+contentDispositionAttachment+"\n"+"\n"
			      storeLog(connection.RemoteAddr().String()+" GET "+file+" 200")
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
			          err := binary.Write(connection, binary.LittleEndian, data2)
			          if err != nil {
			                fmt.Println("err:", err)
			          }
			        } else {
			          total+=len(data)
			          err2 := binary.Write(connection, binary.LittleEndian, data)
			          if err2 != nil {
			                fmt.Println("err:", err2)
			          }
			        }
		      	}
						openfile.Close()
					}
		      connection.Close()
		    }
			}
		}
  }
}
