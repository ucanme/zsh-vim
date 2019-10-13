package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/qiniu/api.v7/auth/qbox"
	"github.com/qiniu/api.v7/storage"
	"math/rand"
	"io"
	"time"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"sync"
)

type res struct {
	Status    int      `json:"status"`
	Data      []string `json:"data"`
	LessionId string   `json:"lessionId,omitempty"`
	CourseId  string   `json:"courseId,omitempty"`
	Type      string   `json:"type"`
	Message   string   `json:"message"`
}

var wg sync.WaitGroup

const maxUploadSize = 5 * 1024 * 1024
const pptPath = "/home/pptResource/"

var chanFilePath = make(chan []string,10);

func main() {
	//处理文件上传
	go convertHandler();

	//提供http服务
	http.HandleFunc("/convert", ConvertHandle)
	fmt.Println("服务启动")
	err := http.ListenAndServe(":9091", nil)
	if err != nil {
		panic(err)
	}
	fmt.Println("Hello Go")
}


/**
 * 离线转化文件格式并上传到七牛云
 */
func convertHandler()  {
	for {
		listInfo := <- chanFilePath
		fmt.Println("-----",listInfo)
		filePath := listInfo[0];
		courseId := listInfo[1];
		lessionId := listInfo[2];
		fileType := listInfo[3];

		rand.Seed(time.Now().Unix())
		randNum :=  rand.Intn(100000)
		num := strconv.Itoa(randNum)
		size := convert(filePath, courseId, lessionId, fileType,strconv.Itoa(randNum))
		fmt.Println("size",size)
		var ReqData []string;
		if size > 0 {
			for i := 0; i < size; i++ {
				ReqData = append(ReqData, "http://xnjzxyimages.bcreat.com/"+fileType+"-"+courseId+"-"+lessionId+"-"+num+"-"+strconv.Itoa(i)+".jpg")
			}
			//成功请求
			HttpPost3(200, courseId, lessionId, ReqData, fileType)
		} else {
			//失败通知
			HttpPost3(10086, courseId, lessionId, ReqData, fileType)
		}
		fmt.Println("Hello World")
	}

}



/**
 * 文件转化
 */
func convert(filePath string, courseId string, lessionId string,fileType string,num string) int{
	fmt.Println("filePath","_____",filePath)
	cmd := exec.Command("libreoffice6.3", "--headless", "--invisible", "--convert-to", "pdf", filePath, "--outdir", "./")
	fmt.Println(cmd)
	err := cmd.Run()
	if err != nil {
		panic(err)
	}
	fmt.Println(filePath)
	//cmd = exec.Command("convert", "-resize", "1200x", "-density", "120", "quality", "100", courseId+"-"+lessionId+"-"+fileType+".pdf", fileType+"-"+courseId+"-"+lessionId+".jpg")
	cmd = exec.Command("convert",  "-density", "120", "quality", "100", courseId+"-"+lessionId+"-"+fileType+".pdf", fileType+"-"+courseId+"-"+lessionId+"-"+num+".jpg")

	err = cmd.Run()
	fmt.Println("pptPath----",pptPath)
	dir_list, e := ioutil.ReadDir(pptPath)
	fmt.Println(dir_list)
	if e != nil {
		return -1;
	}
	size := 0;
	for _, v := range dir_list {
		fileName := v.Name()
		fmt.Println(fileName)
		if strings.Contains(fileName, fileType+"-"+courseId+"-"+lessionId+"-"+num) && strings.Contains(fileName, "jpg") {
			size++;
			wg.Add(1)
			go upload(pptPath+fileName, fileName, fileType)
		}
	}
	wg.Wait()
	return size;
}


/**
 * 处理请求
 */
func ConvertHandle(w http.ResponseWriter, r *http.Request) {
	fmt.Println("http收到请求")
	header := w.Header()
	header.Add("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	courseId := r.FormValue("courseId")
	lessionId := r.FormValue("lessionId")
	fileType := r.FormValue("type")
	Res := res{200, []string{}, lessionId, courseId, fileType, ""}
	boolRet,strRet := handleFile(w, r, courseId, lessionId,fileType)
	if  boolRet == false{
		Res.Status = 10086;
		Res.Message = strRet;
	}else {
		Res.Message = "上传成功";
		chanFilePath<- []string{
			strRet ,courseId ,lessionId ,fileType}
	}
	response, _ := json.Marshal(Res)
	Res.Data = []string{}
	w.Write(response)
}



/**
 * 处理上传文件
 */
func handleFile(w http.ResponseWriter, r *http.Request, courseId string, lessionId string,fileType string) (bool,string) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		return false,"文件过大或者没有上传文件";
	}
	file, header, err := r.FormFile("uploadFile")
	if err != nil {
		return false,"请选择对应的ppt文件";
	}
	defer file.Close()
	fileSuffix := path.Ext(header.Filename) //获取文件后缀
	if fileSuffix != ".pptx" {
		return false,"请选择对应的pptx后缀文件";
	}

	newFileName := "./" + courseId + "-" + lessionId +"-"+fileType+ fileSuffix
	cur, err := os.Create(newFileName)
	defer cur.Close()
	if err != nil {
		return false,"文件上传系统失败";
	}
	io.Copy(cur, file)
	return true,newFileName;
}

func upload(localFile string, uploadedName string, fileType string) {
	fmt.Println("uploadedName--------------",uploadedName);
	type MyPutRet struct {
		Key    string
		Hash   string
		Fsize  int
		Bucket string
		Name   string
	}
	bucket := "images"
	key := uploadedName
	putPolicy := storage.PutPolicy{
		Scope:      bucket,
		ReturnBody: `{"key":"$(key)","hash":"$(etag)","fsize":$(fsize),"bucket":"$(bucket)","name":"$(x:name)"}`,
	}
	accessKey := "6rT0xUXy8yxCUqfeQoETc0xilXDb2fTGQVDZ4QY0"
	secretKey := "3w6u1EGLS2PJdCCQmssfXsMv3PQf9gDlvfw4dt18"
	mac := qbox.NewMac(accessKey, secretKey)
	upToken := putPolicy.UploadToken(mac)
	cfg := storage.Config{}
	formUploader := storage.NewFormUploader(&cfg)
	ret := MyPutRet{}
	putExtra := storage.PutExtra{
		Params: map[string]string{
			"x:name": ""},
	}
	err := formUploader.PutFile(context.Background(), &ret, upToken, key, localFile, &putExtra)
	fmt.Println(ret)
	if err != nil {
		fmt.Println(err)
		wg.Done()
		panic(err)
	}
	fmt.Println("localFile--------",localFile)
	//err = os.Remove(localFile)
	//if err != nil {
	//	panic("delete fail")
	//}
	wg.Done()
}

func HttpPost3(status int, courseId string, lessionId string, imgUrlList []string, fileType string) {
	str := strings.Replace(strings.Trim(fmt.Sprint(imgUrlList), "[]"), " ", ",", -1)
	my_url := "http://xnjzxy.bcreat.com/wechat/remote_order/accept_data"
	fmt.Println(str)
	fmt.Println(url.Values{"status": {strconv.Itoa(status)}, "courseId": {courseId}, "lessionId": {lessionId}, "imgUrl": {str}}.Encode())
	req, err := http.PostForm(my_url, url.Values{"status": {strconv.Itoa(status)}, "courseId": {courseId}, "lessionId": {lessionId}, "imgUrl": {str}, "type": {fileType}})

	if err != nil {
		fmt.Println(courseId + "-" + lessionId + "upload fail")
		fmt.Println(err)
		panic(err)
	}
	defer req.Body.Close()
	body, err := ioutil.ReadAll(req.Body)
	fmt.Println("||||||||||||||||||||||||||||||||", string(body))
	if err != nil {
		fmt.Println(courseId + "-" + lessionId + "upload fail")
		fmt.Println(string(body))
		panic(err)
	}
	fmt.Println(courseId + "-" + lessionId + "upload success")
}

