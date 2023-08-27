package bcutasr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"slices"

	"github.com/spf13/viper"
)

type BcutAsr struct {
	client *http.Client

	data []byte
	name string
	ext  string

	in_boss_key string
	resource_id string
	upload_id   string
	upload_urls []string
	per_size    int

	Etags []string

	download_url string

	task_id string
}

func New() *BcutAsr {
	return &BcutAsr{
		client: &http.Client{},
	}
}

func (self *BcutAsr) Parse(path string) (res *viper.Viper, err error) {
	base_ := strings.Split(filepath.Base(path), ".")
	if len(base_) != 2 {
		err = fmt.Errorf("there is a problem with the input file path")
		return
	}

	self.name = base_[0]
	self.ext = base_[1]

	if !slices.Contains(SUPPORT_SOUND_FORMAT, self.ext) {
		err = fmt.Errorf("only support %v", SUPPORT_SOUND_FORMAT)
		return
	}

	self.data, err = os.ReadFile(path)
	if err != nil {
		return
	}

	err = self.upload()
	if err != nil {
		return
	}

	err = self.uploadPart()
	if err != nil {
		return
	}

	err = self.commitUpload()
	if err != nil {
		return
	}

	err = self.createTask()
	if err != nil {
		return
	}

loop:
	for range time.Tick(time.Millisecond * 300) {
		res, err = self.getResult()
		if err != nil {
			return
		}

		switch res.GetInt("data.state") {
		case ResultStateStop:
			fmt.Println("waiting start")
		case ResultStateRuning:
			fmt.Printf("running-%s\n", res.GetString("data.remark"))
		case ResultStateError:
			err = fmt.Errorf("识别失败: %s", res.GetString("data.remark"))
			break loop
		case ResultStateComplete:
			fmt.Println("complete")
			break loop
		}
	}

	return
}

/*
{"code":0,"message":"0","ttl":1,"data":{"resource_id":"AIA475849201951978579","title":"1","type":2,"in_boss_key":"biz376761249775503627/AIA475849201951978579.MP3","size":11617274,"upload_urls":["http://jssz-boss.biliapi.net/rubick/biz376761249775503627/AIA475849201951978579.MP3?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=ef42ae0361985232%2F20230827%2Fjssz%2Fs3%2Faws4_request\u0026X-Amz-Date=20230827T173706Z\u0026X-Amz-Expires=86400\u0026X-Amz-SignedHeaders=content-length%3Bhost\u0026partNumber=1\u0026uploadId=99fc8c0bfaa01d26\u0026X-Amz-Signature=5e605c2480e7b3379cf3fc00ba6e91aaed26caea9210ec5d0dc825cf22f71cdd","http://jssz-boss.biliapi.net/rubick/biz376761249775503627/AIA475849201951978579.MP3?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=ef42ae0361985232%2F20230827%2Fjssz%2Fs3%2Faws4_request\u0026X-Amz-Date=20230827T173706Z\u0026X-Amz-Expires=86400\u0026X-Amz-SignedHeaders=content-length%3Bhost\u0026partNumber=2\u0026uploadId=99fc8c0bfaa01d26\u0026X-Amz-Signature=e54fb3fc4f68e97a4dcaf09d
fafa1639861293993738b6f00d7db30bfb9abc7d","http://jssz-boss.biliapi.net/rubick/biz376761249775503627/AIA475849201951978579.MP3?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=ef42ae0361985232%2F20230827%2Fjssz%2Fs3%2Faws4_request\u0026X-Amz-Date=20230827T173706Z\u0026X-Amz-Expires=86400\u0026X-Amz-SignedHeaders=content-length%3Bhost\u0026partNumber=3\u0026uploadId=99fc8c0bfaa01d26\u0026X-Amz-Signature=dfc2dd58624f411d306ef5aa23cb55e1209cf8ef15b8fb075a01296aefe58657"],"upload_id":"99fc8c0bfaa01d26","per_size":5242880}}
*/
func (self *BcutAsr) upload() (err error) {
	body := &bytes.Buffer{}
	mp := multipart.NewWriter(body)

	mp.WriteField("type", "2")
	mp.WriteField("name", self.name)
	mp.WriteField("size", strconv.Itoa(len(self.data)))
	mp.WriteField("resource_file_type", self.ext)
	mp.WriteField("model_id", "7")
	err = mp.Close()
	if err != nil {
		return
	}

	var req *http.Request
	req, err = http.NewRequest("POST", API_REQ_UPLOAD, body)
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", mp.FormDataContentType())

	var resp *http.Response
	resp, err = self.client.Do(req)
	if err != nil {
		return
	}

	defer resp.Body.Close()

	res := viper.New()
	res.SetConfigType("json")
	res.ReadConfig(resp.Body)

	if !slices.Contains(res.AllKeys(), "code") {
		err = fmt.Errorf("failed to create upload request")
		return
	}

	if res.GetInt("code") != 0 {
		err = fmt.Errorf(res.GetString("message"))
		return
	}

	self.in_boss_key = res.GetString("data.in_boss_key")
	self.resource_id = res.GetString("data.resource_id")
	self.upload_id = res.GetString("data.upload_id")
	self.upload_urls = res.GetStringSlice("data.upload_urls")
	self.per_size = res.GetInt("data.per_size")

	return
}

/*
{"Accept-Ranges":["bytes"],"Ali-Cdn-Origin-Error-Code":["endOs,200,0"],"Connection":["keep-alive"],"Content-Length":["0"],"Content-Type":["application/octet-stream"],"Cross-Origin-Resource-Policy":["cross-origin"],"Date":["Sun, 27 Aug 2023 18:06:16 GMT"],"Eagleid":["7518a92916931595750613653e"],"Etag":["ad0619b84b40d21c7247ed40eb25faee"],"Server":["Tengine"],"Timing-Allow-Origin":["*"],"Vary":["Origin"],"Via":["cache22.l2et2[980,0], vcache21.cn5428[1020,0]"],"X-Amz-Request-Id":["64eb90977b8480cf"],"X-Amz-Version-Id":["v1.0.0"]}
*/
func (self *BcutAsr) uploadPart() (err error) {
	for clip, url := range self.upload_urls {
		start := clip * self.per_size
		end := (clip + 1) * self.per_size

		var data []byte
		switch {
		case end > len(self.data):
			data = self.data[start:]
		default:
			data = self.data[start:end]
		}

		var req *http.Request
		req, err = http.NewRequest("PUT", url, bytes.NewReader(data))
		if err != nil {
			return
		}

		var resp *http.Response
		resp, err = self.client.Do(req)
		if err != nil {
			return
		}

		self.Etags = append(self.Etags, resp.Header.Get("Etag"))
	}

	return
}

/*
{"code":0,"message":"0","ttl":1,"data":{"resource_id":"AIA475854070884617174","download_url":"http://jssz-boss.bilibili.co/rubick/biz376761249775503627/AIA475854070884617174.MP3?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=ef42ae0361985232%2F20230827%2Fjssz%2Fs3%2Faws4_request\u0026X-Amz-Date=20230827T182530Z\u0026X-Amz-Expires=86400\u0026X-Amz-SignedHeaders=host\u0026X-Amz-Signature=4922d2c38f14981c4fe64abc938cf93af223a683b89ce3c97433fcb120eb1259"}}
*/
func (self *BcutAsr) commitUpload() (err error) {
	body := &bytes.Buffer{}
	mp := multipart.NewWriter(body)

	mp.WriteField("in_boss_key", self.in_boss_key)
	mp.WriteField("resource_id", self.resource_id)
	mp.WriteField("etags", strings.Join(self.Etags, ","))
	mp.WriteField("upload_id", self.upload_id)
	mp.WriteField("model_id", "7")
	err = mp.Close()
	if err != nil {
		return
	}

	var req *http.Request
	req, err = http.NewRequest("POST", API_COMMIT_UPLOAD, body)
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", mp.FormDataContentType())

	var resp *http.Response
	resp, err = self.client.Do(req)
	if err != nil {
		return
	}

	defer resp.Body.Close()

	res := viper.New()
	res.SetConfigType("json")
	res.ReadConfig(resp.Body)

	if !slices.Contains(res.AllKeys(), "code") {
		err = fmt.Errorf("failed to commit the upload")
		return
	}

	if res.GetInt("code") != 0 {
		err = fmt.Errorf(res.GetString("message"))
		return
	}

	self.download_url = res.GetString("data.download_url")

	return
}

/*
{"code":0,"message":"0","ttl":1,"data":{"resource":"","result":"","task_id":"asr475855047184368511"}}
*/
func (self *BcutAsr) createTask() (err error) {
	data := map[string]interface{}{
		"resource": self.download_url,
		"model_id": "7",
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return
	}

	var req *http.Request
	req, err = http.NewRequest("POST", API_CREATE_TASK, bytes.NewReader(jsonData))
	if err != nil {
		return
	}

	var resp *http.Response
	resp, err = self.client.Do(req)
	if err != nil {
		return
	}

	defer resp.Body.Close()

	res := viper.New()
	res.SetConfigType("json")
	res.ReadConfig(resp.Body)

	if !slices.Contains(res.AllKeys(), "code") {
		err = fmt.Errorf("failed to create the task")
		return
	}

	if res.GetInt("code") != 0 {
		err = fmt.Errorf(res.GetString("message"))
		return
	}

	self.task_id = res.GetString("data.task_id")

	return
}

/*
{"code":0,"message":"0","ttl":1,"data":{"task_id":"asr475855435308488383","result":"","remark":"","state":0}}

result.data
*/
func (self *BcutAsr) getResult() (res *viper.Viper, err error) {
	params := map[string]string{
		"model_id": "7",
		"task_id":  self.task_id,
	}

	var req *http.Request
	req, err = http.NewRequest("GET", API_QUERY_RESULT, nil)
	if err != nil {
		return
	}

	q := req.URL.Query()
	for key, value := range params {
		q.Add(key, value)
	}
	req.URL.RawQuery = q.Encode()

	var resp *http.Response
	resp, err = self.client.Do(req)
	if err != nil {
		return
	}

	defer resp.Body.Close()

	res = viper.New()
	res.SetConfigType("json")
	res.ReadConfig(resp.Body)

	if !slices.Contains(res.AllKeys(), "code") {
		err = fmt.Errorf("failed to create the task")
		return
	}

	if res.GetInt("code") != 0 {
		err = fmt.Errorf(res.GetString("message"))
		return
	}

	return
}
