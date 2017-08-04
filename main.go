package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"time"

	"os/exec"

	"github.com/Comdex/imgo"
)

var Config struct {
	NoxPath string // 夜神模拟器路径
	Sleep   int64  // 等待间隔, 单位为秒
}

func init() {
	bytes, err := ioutil.ReadFile("./config.json")
	if err != nil {
		log.Fatalln(err)
	}
	if err := json.Unmarshal(bytes, &Config); err != nil {
		log.Fatalln(err)
	}
}

var device string

func main() {
	if b, err := exec.Command(Config.NoxPath+"nox_adb.exe", "devices").Output(); err != nil {
		log.Fatalln(err)
	} else {
		fmt.Println(string(b))
	}
	fmt.Println("Please input your device serial number: ")
	for device == "" {
		fmt.Scanln(&device)
	}
	if err := exec.Command(Config.NoxPath+"nox_adb.exe", "-s", device,
		"shell", "ls",
	).Run(); err != nil {
		log.Fatalln(err)
	}
	fmt.Println("start-up success")
	for {
	TAT:
		getScreenshot()
		if cosineSimilarity(imgo.MustRead("./tpl/t0.png"), getT("./screenshot.png")) >= 0.98 {
			log.Println("你的老婆回家啦 ヾ(•ω•`)o")
			tap(450, 640) // 随便点个地方, 收回后勤支援
			time.Sleep(time.Second * 1)
			tap(743, 492) // 点击确定
			time.Sleep(time.Second * 1)
			tap(500, 200) // 摸摸老婆的头
			time.Sleep(time.Second * 5)
			goto TAT
		}
		os.Remove("./screenshot.png")
		time.Sleep(time.Second * time.Duration(Config.Sleep))
	}
}

// 两图余弦相似度
func cosineSimilarity(matrix1 [][][]uint8,
	matrix2 [][][]uint8) (cossimi float64) {
	myx := imgo.Matrix2Vector(matrix1)
	myy := imgo.Matrix2Vector(matrix2)
	cos1 := imgo.Dot(myx, myy)
	cos21 := math.Sqrt(imgo.Dot(myx, myx))
	cos22 := math.Sqrt(imgo.Dot(myy, myy))
	cossimi = cos1 / (cos21 * cos22)
	return
}

// 截取模拟器快照
func getScreenshot() {
	if err := exec.Command(Config.NoxPath+"nox_adb.exe", "-s", device,
		"shell", "/system/bin/screencap", "-p", "/sdcard/screenshot.png",
	).Run(); err != nil {
		log.Fatalln(err)
	}

	wd, err := os.Getwd()
	if err != nil {
		log.Fatalln(err)
	}
	if err := exec.Command(Config.NoxPath+"nox_adb.exe", "-s", device,
		"pull", "/sdcard/screenshot.png", wd+`/screenshot.png`,
	).Run(); err != nil {
		log.Fatalln(err)
	}
}

// 点击模拟器
func tap(x int64, y int64) {
	if err := exec.Command(Config.NoxPath+"nox_adb.exe", "-s", device,
		"shell", "input", "tap", fmt.Sprint(x), fmt.Sprint(y),
	).Run(); err != nil {
		log.Fatalln(err)
	}
}

// 截取模板
func getT(fileName string) (img [][][]uint8) {
	img = imgo.New3DSlice(67, 71, 4)
	s := imgo.MustRead(fileName)
	for i := 0; i < 67; i++ {
		for h := 0; h < 71; h++ {
			img[i][h] = append([]uint8{}, s[1213+i][649+h]...)
		}
	}
	return
}
