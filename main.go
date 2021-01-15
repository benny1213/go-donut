package main

import (
	"fmt"
	"math"
	"syscall"
	"time"

	"golang.org/x/sys/unix"
)

// var ws *unix.Winsize
type wsType struct {
	*unix.Winsize
	xRatio float64
	yRatio float64
}

var ws wsType

const thetaSpacing float64 = 0.07
const phiSpacing float64 = 0.02
const r1 float64 = 1
const r2 float64 = 2
const k2 float64 = 5

// 显示在3/8屏幕宽度的大小上
// (r1 + r2) / k2 = (screenWidth * 3/8) / k1
var k1 float64

func init() {
	setWs()
}

func setWs() {
	var err error
	winSize, err := unix.IoctlGetWinsize(syscall.Stdout, unix.TIOCGWINSZ)
	if err != nil {
		panic(err)
	}
	ws = wsType{
		xRatio:  float64(winSize.Col) / float64(winSize.Xpixel),
		yRatio:  float64(winSize.Row) / float64(winSize.Ypixel),
		Winsize: winSize,
	}
	k1 = float64(ws.Ypixel) * k2 * 3 / (8 * (r1 + r2))
}

func main() {
	var A float64 = 0
	var B float64 = 90
	// renderFrame(A, B)
	for {
		renderFrame(A, B)
		time.Sleep(time.Duration(50) * time.Millisecond)
		A += 0.1
		B += 0.1
	}

	// renderCercle()
}

func renderFrame(A, B float64) {
	setWs()
	// 初始化二维数组
	output := make([][]rune, int(ws.Col))
	for i := range output {
		output[i] = make([]rune, int(ws.Row))
	}
	zbuffer := make([][]float64, int(ws.Col))
	for i := range output {
		zbuffer[i] = make([]float64, int(ws.Row))
	}

	// 计算点位置
	cosA := math.Cos(A)
	sinA := math.Sin(A)
	cosB := math.Cos(B)
	sinB := math.Sin(B)
	for theta := 0.0; theta < 2*math.Pi; theta += thetaSpacing {
		cosTheta := math.Cos(theta)
		sinTheta := math.Sin(theta)
		for phi := 0.0; phi < 2*math.Pi; phi += phiSpacing {
			cosPhi := math.Cos(phi)
			sinPhi := math.Sin(phi)

			// 旋转前的圆
			circleX := r2 + r1*cosTheta
			circleY := r1 * sinTheta

			// 经过旋转矩阵计算后的坐标
			x := circleX*(cosB*cosPhi+sinA*sinB*sinPhi) - cosA*sinB*circleY
			y := circleX*(cosPhi*sinB-cosB*sinA*sinPhi) + cosA*cosB*circleY
			z := k2 + cosA*circleX*sinPhi + circleY*sinA

			ooz := 1 / z

			xp := int((float64(ws.Xpixel)/2 + k1*ooz*x) * ws.xRatio)
			yp := int((float64(ws.Ypixel)/2-k1*ooz*y)*ws.yRatio + 1)
			// 计算流明
			L := cosPhi*cosTheta*sinB - cosA*cosTheta*sinPhi - sinA*sinTheta + cosB*(cosA*sinTheta-cosTheta*sinA*sinPhi)
			if L > 0 {
				if xp > 0 &&
					xp < int(ws.Col) &&
					yp > 0 &&
					yp < int(ws.Row) &&
					ooz > zbuffer[xp][yp] {

					zbuffer[xp][yp] = ooz
					luminanceIndex := int(L * 8)
					output[xp][yp] = []rune(".,-~:;=!*#$@")[luminanceIndex]
				}
			}
		}
	}
	// 打印
	fmt.Print("\x1b[H")
	for j := 0; j < int(ws.Row); j++ {
		for i := 0; i < int(ws.Col); i++ {
			if output[i][j] == 0 {
				fmt.Printf(" ")
			} else {
				fmt.Printf("%c", output[i][j])
			}
		}
		fmt.Print("\n")
	}
}
