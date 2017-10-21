package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"regexp"
	"strconv"
	"time"

	"os/exec"

	"github.com/Comdex/imgo"
)

var (
	Config struct {
		AdbPath string // adb路径
		Sleep   int64  // 等待间隔, 单位为秒
	}

	app struct {
		device string
		width  int
		height int
		dpi    int
		tpl    string
	}
)

func init() {
	bytes, err := ioutil.ReadFile("./config.json")
	if err != nil {
		log.Fatalln(err)
	}
	if err := json.Unmarshal(bytes, &Config); err != nil {
		log.Fatalln(err)
	}
}

func main() {
	if b, err := exec.Command(Config.AdbPath, "devices").Output(); err != nil {
		log.Fatalln(err)
	} else {
		fmt.Println(string(b))
	}
	fmt.Println("Please input your device serial number: ")
	for app.device == "" {
		fmt.Scanln(&app.device)
	}

	app.width, app.height, app.dpi = getSize()
	if app.dpi != 240 {
		panic("device dpi not 240")
	} else if app.width == 0 || app.height == 0 {
		panic("device width or height error")
	}
	app.tpl = fmt.Sprintf("./tpl/%vx%v.png", app.width, app.height)
	fmt.Println("start-up success")

	for {
	TAT:
		getScreenshot()
		if cosineSimilarity(imgo.MustRead(app.tpl), getTpl("./screenshot.png")) >= 0.98 {
			log.Println("go home")
			tap(1, app.height-1) // 随便点个地方, 收回后勤支援
			time.Sleep(time.Second * 1)
			tap(743, 492) // 点击确定
			time.Sleep(time.Second * 1)
			tap(500, 200) // 摸摸头
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
	if err := exec.Command(Config.AdbPath, "-s", app.device,
		"shell", "/system/bin/screencap", "-p", "/sdcard/screenshot.png",
	).Run(); err != nil {
		log.Fatalln(err)
	}

	wd, err := os.Getwd()
	if err != nil {
		log.Fatalln(err)
	}
	if err := exec.Command(Config.AdbPath, "-s", app.device,
		"pull", "/sdcard/screenshot.png", wd+`/screenshot.png`,
	).Run(); err != nil {
		log.Fatalln(err)
	}
}

// 点击设备
func tap(x int, y int) {
	if err := exec.Command(Config.AdbPath, "-s", app.device,
		"shell", "input", "tap", fmt.Sprint(x), fmt.Sprint(y),
	).Run(); err != nil {
		log.Fatalln(err)
	}
}

// 截取模板
func getTpl(fileName string) (img [][][]uint8) {
	img = imgo.New3DSlice(50, 50, 4)
	s := imgo.MustRead(fileName)
	_height := len(s)
	_width := 0
	if _height > 0 {
		_width = len(s[0])
	}
	if _height < 50 || _width < 50 {
		return
	}
	for i := 0; i < 50; i++ {
		_h := i
		if _width < _height {
			_h = _height - 50 + i
		}
		for h := 0; h < 50; h++ {
			_w := _width - 50 + h
			img[i][h] = append([]uint8{}, s[_h][_w]...)
		}
	}
	return
}

// 获取设备分辨率
func getSize() (_width, _height, _dpi int) {
	// adb shell dumpsys window displays
	b, err := exec.Command(Config.AdbPath, "-s", app.device,
		"shell", "dumpsys", "window", "displays").Output()
	if err != nil {
		panic(err)
	}
	r, _ := regexp.Compile(`init=([0-9]{3,})x([0-9]{3,}) ([0-9]*)dpi`)
	_size := r.FindAllStringSubmatch(string(b), -1)
	if len(_size) == 1 && len(_size[0]) == 4 {
		_width, _ = strconv.Atoi(_size[0][1])
		_height, _ = strconv.Atoi(_size[0][2])
		_dpi, _ = strconv.Atoi(_size[0][3])
	}
	return
}
