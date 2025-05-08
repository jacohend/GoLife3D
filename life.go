// main.go
package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"runtime"
	"sync"
	"time"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

const (
	SCREEN_WIDTH  = 1000 // 10 000 px windows often fail on modern drivers
	SCREEN_HEIGHT = 1000
	SIZE          = 100 // voxel grid edge length
)

var (
	window                 *glfw.Window
	X_AXIS, Y_AXIS, Z_AXIS float32
	DIRECTION              int
	universe               [SIZE][SIZE][SIZE]uint8
	mu                     sync.Mutex
)

func fatalIf(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func resizeWindow(width, height int) {
	if height == 0 {
		height = 1
	}
	gl.Viewport(0, 0, int32(width), int32(height))

	gl.MatrixMode(gl.PROJECTION)
	gl.LoadIdentity()

	aspect := float64(width) / float64(height)
	fov, near, far := 15.0, 0.1, 100.0
	top := math.Tan(fov*math.Pi/360.0) * near
	bottom := -top
	left := aspect * bottom
	right := aspect * top
	gl.Frustum(left, right, bottom, top, near, far)

	gl.MatrixMode(gl.MODELVIEW)
	gl.LoadIdentity()
}

// --- drawing helpers -------------------------------------------------------

func draw(i, j, k int) {
	n := float32(SIZE)
	newi := (1 + float32(i) - n/2) / n
	newj := (1 + float32(j) - n/2) / n
	newk := (1 + float32(k) - n/2) / n

	gl.Color3f(float32(i)/n, float32(j)/n, float32(k)/n)
	gl.Vertex3f(newi, newj, newk)
}

func drawGLScene() {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	gl.LoadIdentity()

	gl.Translatef(0, 0, -6)
	gl.Rotatef(X_AXIS, 1, 0, 0)
	gl.Rotatef(Y_AXIS, 0, 1, 0)
	gl.Rotatef(Z_AXIS, 0, 0, 1)

	gl.Begin(gl.POINTS)
	mu.Lock()
	for x := 0; x < SIZE; x++ {
		for y := 0; y < SIZE; y++ {
			for z := 0; z < SIZE; z++ {
				if universe[x][y][z] == 1 {
					draw(x, y, z)
				}
			}
		}
	}
	mu.Unlock()
	gl.End()

	// gentle spin
	for _, axis := range []*float32{&X_AXIS, &Y_AXIS, &Z_AXIS} {
		if *axis > 360 || *axis < -360 {
			*axis = 0
		}
		*axis += 0.3 * float32(DIRECTION)
	}
}

// --- input / callbacks ------------------------------------------------------

func handleKeyPress(key glfw.Key) {
	switch key {
	case glfw.KeyEscape:
		window.SetShouldClose(true)
	case glfw.KeyA:
		DIRECTION = -1
		Y_AXIS -= 0.30
	case glfw.KeyS:
		DIRECTION = 1
		X_AXIS += 0.30
	case glfw.KeyD:
		DIRECTION = 1
		Y_AXIS += 0.30
	case glfw.KeyW:
		DIRECTION = -1
		X_AXIS -= 0.30
	case glfw.KeyQ:
		DIRECTION = 0
	}
}

// --- simulation -------------------------------------------------------------

func createUniverse() {
	rand.Seed(time.Now().UnixNano())
	for x := 0; x < SIZE; x++ {
		for y := 0; y < SIZE; y++ {
			for z := 0; z < SIZE; z++ {
				universe[x][y][z] = uint8(rand.Intn(2))
			}
		}
	}
}

func bruteNeighbors(i, j, k int) int {
	neighbors := 0
	for x := i - 1; x <= i+1; x++ {
		for y := j - 1; y <= j+1; y++ {
			for z := k - 1; z <= k+1; z++ {
				if x == i && y == j && z == k {
					continue
				}
				if x >= 0 && x < SIZE &&
					y >= 0 && y < SIZE &&
					z >= 0 && z < SIZE &&
					universe[x][y][z] == 1 {
					neighbors++
				}
			}
		}
	}
	return neighbors
}

func simulate() {
	temp := universe
	for {
		var next [SIZE][SIZE][SIZE]uint8
		for x := 0; x < SIZE; x++ {
			for y := 0; y < SIZE; y++ {
				for z := 0; z < SIZE; z++ {
					cell := temp[x][y][z]
					n := bruteNeighbors(x, y, z)
					if cell == 1 && (n < 4 || n > 5) {
						next[x][y][z] = 0
					} else if cell == 0 && (n == 4 || n == 5) {
						next[x][y][z] = 1
					} else {
						next[x][y][z] = cell
					}
				}
			}
		}
		mu.Lock()
		universe = next
		mu.Unlock()
		temp = next
		time.Sleep(500 * time.Millisecond)
	}
}

// --- main -------------------------------------------------------------------

func initGL() {
	gl.ShadeModel(gl.SMOOTH)
	gl.ClearColor(0, 0, 0, 0)
	gl.ClearDepth(1)
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)
	gl.Hint(gl.PERSPECTIVE_CORRECTION_HINT, gl.NICEST)
}

func main() {
	// OpenGL expects a single OS thread
	runtime.LockOSThread()

	fatalIf(glfw.Init())
	defer glfw.Terminate()

	glfw.WindowHint(glfw.Resizable, glfw.True)
	glfw.WindowHint(glfw.ContextVersionMajor, 2)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)

	var err error
	window, err = glfw.CreateWindow(SCREEN_WIDTH, SCREEN_HEIGHT, "Voxel Universe", nil, nil)
	fatalIf(err)
	window.MakeContextCurrent()
	glfw.SwapInterval(1) // vsync

	fatalIf(gl.Init())
	fmt.Println("OpenGL version:", gl.GoStr(gl.GetString(gl.VERSION)))

	window.SetFramebufferSizeCallback(func(w *glfw.Window, width, height int) {
		resizeWindow(width, height)
	})
	window.SetKeyCallback(func(w *glfw.Window, key glfw.Key, sc int, action glfw.Action, mods glfw.ModifierKey) {
		if action == glfw.Press {
			handleKeyPress(key)
		}
	})

	initGL()
	resizeWindow(SCREEN_WIDTH, SCREEN_HEIGHT)

	createUniverse()
	go simulate()

	for !window.ShouldClose() {
		drawGLScene()
		window.SwapBuffers()
		glfw.PollEvents()
	}
}
