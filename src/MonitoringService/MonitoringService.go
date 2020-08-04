package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/tidwall/gjson"
	//"reflect"
	client "github.com/influxdata/influxdb1-client/v2"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/debug"
	//"golang.org/x/sys/windows/svc/debug"
)

// Json 구조체 선언
type BaseStruct struct {
	MainServerStatus string
	Etc1             int64
	Etc2             int64
	Etc3             int64
	Etc4             int64
	ServerID         int64
	UserCount        int64
	LoginCount       int64
}

const (
	MyDB     = "test1"
	username = "user"
	password = "pw"
)

var wg sync.WaitGroup

/*
//env.json 파일 변수 정보
env_data_result["InfluxDB_Address"]
env_data_result["Interval"]
env_data_result["Stored_DB"]
env_data_result["User"]
env_data_result["Password"]
*/

func errCheck(e error) {
	if e != nil {
		panic(e)
	}
}

func insertData(gameid string, url string) {

	env_data, err := ioutil.ReadFile("env.json") //--> byte로 받아 오기 때문에 String으로 형 변황을 해줘야 함
	errCheck(err)
	var env_data_result map[string]interface{}
	json.Unmarshal([]byte(env_data), &env_data_result)

	var doc string
	response, err := http.Get(url)
	if err != nil {
		fmt.Printf("%s", err)
		os.Exit(1)
	} else {
		defer response.Body.Close()
		contents, err := ioutil.ReadAll(response.Body)
		if err != nil {
			fmt.Printf("%s", err)
			os.Exit(1)
		}

		doc = string(contents)

	}

	//declare valiable
	var mainServerStatus string
	var etc1 string
	var etc2 string
	var etc3 string
	var etc4 string
	var serverID string
	var userCount string
	var loginCount string

	data := make(map[string]BaseStruct)

	//convert b, gjson.result to int64

	//Json 파일의 Line Count를 가져온다.
	a := gjson.Get(doc, "serverList.#")
	stringB := a.String()
	linecount, err := strconv.Atoi(stringB)
	if err == nil {
		fmt.Println(linecount)
	}

	//define influxdb connection
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     env_data_result["InfluxDB_Address"].(string),
		Username: env_data_result["User"].(string),
		Password: env_data_result["Password"].(string),
	})
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	// end influxdb conntion

	for i := 0; i <= linecount; i++ {
		mainServerStatus = "serverList." + strconv.Itoa(i) + ".logicalServerStatus.mainServerStatus"
		etc1 = "serverList." + strconv.Itoa(i) + ".etc.etc1"
		etc2 = "serverList." + strconv.Itoa(i) + ".etc.etc2"
		etc3 = "serverList." + strconv.Itoa(i) + ".etc.etc3"
		etc4 = "serverList." + strconv.Itoa(i) + ".etc.etc4"
		serverID = "serverList." + strconv.Itoa(i) + ".serverId"
		userCount = "serverList." + strconv.Itoa(i) + ".userCount"
		loginCount = "serverList." + strconv.Itoa(i) + ".loginCount"

		//Get Json
		serverID := gjson.Get(doc, serverID)                 //gjson Type
		mainServerStatus := gjson.Get(doc, mainServerStatus) //gjson Type
		etc1 := gjson.Get(doc, etc1)                         //gjson Type
		etc2 := gjson.Get(doc, etc2)                         //gjson Type
		etc3 := gjson.Get(doc, etc3)                         //gjson Type
		etc4 := gjson.Get(doc, etc4)                         //gjson Type
		userCount := gjson.Get(doc, userCount)               //gjson Type
		loginCount := gjson.Get(doc, loginCount)             //gjson Type

		//fmt.Println(reflect.TypeOf(ServerID))
		//fmt.Println(reflect.TypeOf(ServerID.Int()))
		//fmt.Println(ServerID.String())
		//fmt.Println(MainServerStatus.String())
		if serverID.Int() != 0 {
			data[strconv.Itoa(i)] = BaseStruct{
				ServerID:         serverID.Int(),
				MainServerStatus: mainServerStatus.String(),
				Etc1:             etc1.Int(),
				Etc2:             etc2.Int(),
				Etc3:             etc3.Int(),
				Etc4:             etc4.Int(),
				UserCount:        userCount.Int(),
				LoginCount:       loginCount.Int()}
		}

		/*
			if serverID.Int() != 0 {
				fmt.Println(data[strconv.Itoa(i)].ServerID, data[strconv.Itoa(i)].MainServerStatus, data[strconv.Itoa(i)].Etc1, data[strconv.Itoa(i)].Etc2, data[strconv.Itoa(i)].Etc3, data[strconv.Itoa(i)].UserCount, data[strconv.Itoa(i)].LoginCount)
		}*/

		//create batch points
		bp, err := client.NewBatchPoints(client.BatchPointsConfig{
			Database:  env_data_result["Stored_DB"].(string),
			Precision: "ms",
		})
		if err != nil {
			log.Fatal(err)
		}
		// Create a point and add to batch
		if serverID.Int() != 0 {
			tags := map[string]string{"ServerID": strconv.FormatInt(data[strconv.Itoa(i)].ServerID, 10)}
			//var test1 int64
			//var test2 string
			//test1 = data[strconv.Itoa(i)].ServerID
			//test2 = strconv.FormatInt(test1, 10)
			//fmt.Println(test1, test2)

			//tags := map[string]int{"ServerID": data.ServerID}
			fields := map[string]interface{}{
				"MainServerStatus": data[strconv.Itoa(i)].MainServerStatus,
				"Etc1":             data[strconv.Itoa(i)].Etc1,
				"Etc2":             data[strconv.Itoa(i)].Etc2,
				"Etc3":             data[strconv.Itoa(i)].Etc3,
				"Etc4":             data[strconv.Itoa(i)].Etc4,
				"UserCount":        data[strconv.Itoa(i)].UserCount,
				"LoginCount":       data[strconv.Itoa(i)].LoginCount,
			}

			pt, err := client.NewPoint(gameid, tags, fields, time.Now())
			if err != nil {
				log.Fatal(err)
			}

			bp.AddPoint(pt)

			// Write the batch
			if err := c.Write(bp); err != nil {
				log.Fatal(err)
			}
		}
		// Close client resources
		if err := c.Close(); err != nil {
			log.Fatal(err)
		}
	}
	wg.Done()
}

// 서비스 Type
type dummyService struct {
}

// svc.Handler 인터페이스 구현
func (srv *dummyService) Execute(args []string, req <-chan svc.ChangeRequest, stat chan<- svc.Status) (svcSpecificEC bool, exitCode uint32) {
	stat <- svc.Status{State: svc.StartPending}

	// 실제 서비스 내용
	stopChan := make(chan bool, 1)
	go runBody(stopChan)

	stat <- svc.Status{State: svc.Running, Accepts: svc.AcceptStop | svc.AcceptShutdown}

LOOP:
	for {
		// 서비스 변경 요청에 대해 핸들링
		switch r := <-req; r.Cmd {
		case svc.Stop, svc.Shutdown:
			stopChan <- true
			break LOOP

		case svc.Interrogate:
			stat <- r.CurrentStatus
			time.Sleep(100 * time.Millisecond)
			stat <- r.CurrentStatus

			//case svc.Pause:
			//case svc.Continue:
		}
	}

	stat <- svc.Status{State: svc.StopPending}
	return
}

/*** 서비스에서 실제 하는 일 ***/
func runBody(stopChan chan bool) {
	config_data, err := ioutil.ReadFile("config.json") //--> byte로 받아 오기 때문에 String으로 형 변황을 해줘야 함
	errCheck(err)
	var config_data_result map[string]interface{}
	json.Unmarshal([]byte(config_data), &config_data_result)
	env_data, err := ioutil.ReadFile("env.json") //--> byte로 받아 오기 때문에 String으로 형 변황을 해줘야 함
	errCheck(err)
	var env_data_result map[string]interface{}
	json.Unmarshal([]byte(env_data), &env_data_result)

	for {
		select {
		case <-stopChan:
			return
		default:
			// 10초 마다 현재시간 새로 쓰기
			count := env_data_result["Interval"].(float64)
			time.Sleep(time.Duration(count) * time.Second)
			//ioutil.WriteFile("C:/temp/log.txt", []byte(time.Now().String()), 0)

			for k, v := range config_data_result {
				fmt.Println("GetInfo: ", k, v)
				wg.Add(1)
				go insertData(k, v.(string))
			}

			wg.Wait()
			fmt.Println("insert Complete!")
		}
	}
}

func main() {
	//err := svc.Run("DummyService", &dummyService{})
	err := debug.Run("DummyService", &dummyService{}) //콘솔출력 디버깅시
	if err != nil {
		panic(err)
	}

}