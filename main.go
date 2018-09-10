package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
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
		device     string
		width      int
		height     int
		dpi        int
		screenshot chan int
		tap        chan int
	}
)

func init() {
	app.screenshot = make(chan int, 1)
	app.tap = make(chan int, 1)
	bytes, err := ioutil.ReadFile("./config.json")
	if err != nil {
		log.Fatalln(err)
	}
	if err := json.Unmarshal(bytes, &Config); err != nil {
		log.Fatalln(err)
	}
}

func isOK(fileName string) (x, y int) {
	im := imgo.MustRead(fileName)
	h := len(im)
	w := len(im[0])
	if h < 11 || w < 11 ||
		DeepEqual_(im[40][w-385], []uint8{49, 49, 49, 255}) {
		// log.Println(im[10][w-10])
		return -1, -1
	}
	var oldRGBA []uint8
	re := 0
	for x := (h / 2); x < h; x++ {
		for y := 0; y < w; y++ {
			if len(oldRGBA) > 1 {
				if im[x][y][0] == oldRGBA[0] &&
					im[x][y][1] == oldRGBA[1] &&
					im[x][y][2] == oldRGBA[2] {
					re++
				} else {
					re = 0
				}
				if DeepEqual_(im[x][y], []uint8{255, 180, 0, 255}) {
					// log.Println("isOK", im[x][y])
					if DeepEqual_(im[40][w-385], []uint8{49, 49, 49, 255}) {
						return -1, -1
					}
					return x, y
				}
			}
			oldRGBA = im[x][y]
		}
	}
	return -1, -1
}

func isGoHome(fileName string) bool {
	im := imgo.MustRead(fileName)
	h := len(im)
	w := len(im[0])
	var oldRGBA []uint8
	re := 0
	if h < 11 || w < 11 {
		return false
	} else if DeepEqual_(im[31][w-375], []uint8{255, 178, 0, 255}) && DeepEqual_(im[40][w-385], []uint8{49, 49, 49, 255}) {
		return false
	}
	for x := (h / 2); x < h; x++ {
		for y := 0; y < (w / 2); y++ {
			if len(oldRGBA) > 1 {
				if DeepEqual(im[x][y], oldRGBA) &&
					DeepEqual(oldRGBA, []uint8{255, 255, 255, 255}) {
					hp := int(h / 30)
					if x+hp < h {
						for r := 0; r < hp; r++ {
							if !DeepEqual(im[x+r][y], []uint8{255, 255, 255, 255}) {
								if r >= hp-1 {
									re++
								}
								break
							}
						}
					}
				}
			}
			oldRGBA = im[x][y]
		}
		if x >= (h/4)*3 {
			break
		}
	}
	return re >= 10
}

func main() {
	if b, err := exec.Command(Config.AdbPath, "devices").Output(); err != nil {
		log.Fatalln(err)
	} else {
		fmt.Println(string(b))
	}
	fmt.Println("Please input your device serial number: ")
	app.device = "192.168.56.101:5555"
	for app.device == "" {
		fmt.Scanln(&app.device)
	}

	app.width, app.height, app.dpi = getSize()
	fmt.Println("size", app.width, app.height, app.dpi)
	fmt.Println("start-up success")

	go func() {
		for {
			getScreenshot("message")
			x, y := isOK("./screenshot/message.png")
			// log.Println(x, y)
			if x > 0 && y > 0 {
				log.Println("再次出征")
				tap(y, x)
				time.Sleep(time.Second * 1)
				// os.Exit(0)
			}
			os.Remove("./screenshot/message.png")
			time.Sleep(time.Second * time.Duration(Config.Sleep))
		}
	}()
	go func() {
		for {
			getScreenshot("home")
			if isGoHome("./screenshot/home.png") {
				log.Println("后勤完毕")
				tap(app.width-282, app.height-5)
			}
			os.Remove("./screenshot/home.png")
			time.Sleep(time.Second * time.Duration(Config.Sleep))
		}
	}()
	make(chan int) <- 0
}

// 截取模拟器快照
func getScreenshot(fileName string) {
	app.screenshot <- 1
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
		"pull", "/sdcard/screenshot.png", wd+`/screenshot/`+fileName+`.png`,
	).Run(); err != nil {
		log.Fatalln(err)
	}
	<-app.screenshot
}

// 点击设备
func tap(x int, y int) {
	app.tap <- 1
	if err := exec.Command(Config.AdbPath, "-s", app.device,
		"shell", "input", "tap", fmt.Sprint(x), fmt.Sprint(y),
	).Run(); err != nil {
		log.Fatalln(err)
	}
	<-app.tap
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
		a_, _ := strconv.Atoi(_size[0][1])
		b_, _ := strconv.Atoi(_size[0][2])
		if a_ > b_ {
			_height = b_
			_width = a_
		} else {
			_height = a_
			_width = b_
		}
		_dpi, _ = strconv.Atoi(_size[0][3])
	}
	return
}

func DeepEqual_(a, b []uint8) bool {
	if len(a) != len(b) {
		return false
	}
	if (a == nil) != (b == nil) {
		return false
	}
	for i, v := range a {
		if !(b[i] < v+10 && b[i] > v-10 || v == b[i]) {
			return false
		}
		// if v != b[i] {
		// 	return false
		// }
	}
	return true
}

func DeepEqual(a, b []uint8) bool {
	if len(a) != len(b) {
		return false
	}
	if (a == nil) != (b == nil) {
		return false
	}
	for i, v := range a {
		if !(b[i] < v+10 && b[i] > v-10 || v == b[i]) {
			return false
		}
		// if v != b[i] {
		// 	return false
		// }
	}
	return true
}
