package main

import (
    "github.com/go-gl/gl"
    "fmt"
    "math"
    "github.com/banthar/Go-SDL/sdl"
    "os"
    "math/rand"
    "time"
    "runtime"
    "sync"
)

const (
    SCREEN_WIDTH  = 1000
    SCREEN_HEIGHT = 1000
    SCREEN_BPP    = 32
    SIZE = 100
)

var (
    surface *sdl.Surface
    X_AXIS, Y_AXIS, Z_AXIS float32
    DIRECTION int
)


var universe[SIZE][SIZE][SIZE]uint8
var mu sync.Mutex

func Quit(status int) {
    sdl.Quit()
    os.Exit(status)
}

func resizeWindow(width, height int) {
    if height == 0 {
        height = 1
    }
    gl.Viewport(0, 0, width, height)
    gl.MatrixMode(gl.PROJECTION)
    gl.LoadIdentity()
    aspect := gl.GLdouble(width / height)
    var fov, near, far gl.GLdouble
    fov = 15.0
    near = 0.1
    far = 100.0
    top := gl.GLdouble(math.Tan(float64(fov*math.Pi/360.0))) * near
    bottom := -top
    left := aspect * bottom
    right := aspect * top
    gl.Frustum(float64(left), float64(right), float64(bottom), float64(top), float64(near), float64(far))
    gl.MatrixMode(gl.MODELVIEW)
    gl.LoadIdentity()
}

func draw(i, j, k int){
    newi := ((1.0 + (float32(i)+0.0)- (float32(SIZE)/2.0)) /float32(SIZE))
    newj := ((1.0 + (float32(j)+0.0)- (float32(SIZE)/2.0)) /float32(SIZE))
    newk := ((1.0 + (float32(k)+0.0)- (float32(SIZE)/2.0)) /float32(SIZE))
    gl.Color3f(float32(float32(i)/float32(SIZE)),float32(float32(j)/float32(SIZE)),float32(float32(k)/float32(SIZE)))
    gl.Vertex3f(newi, newj, newk)
}

func drawGLScene() {
    gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
    gl.LoadIdentity()
    gl.Translatef(0.0, 0.0, -6.0)
    gl.Rotatef(X_AXIS, 1.0, 0.0, 0.0)
    gl.Rotatef(Y_AXIS, 0.0, 1.0, 0.0)
    gl.Rotatef(Z_AXIS, 0.0, 0.0, 1.0)
    gl.Begin(gl.POINTS)
    gl.Color3f(0.0,0.0,0.0)
    mu.Lock()
    for x := range make([]int, SIZE){
        for y := range make([]int, SIZE){
            for z := range make([]int, SIZE){
                if universe[x][y][z] == 1{
                    draw(x, y, z)
                }
            }
        }
    }
    mu.Unlock()
    gl.End()
    if X_AXIS > 360 || X_AXIS == 0 || X_AXIS < -360{
        X_AXIS = 0.0
    }else{
        X_AXIS = X_AXIS + 0.3*float32(DIRECTION)
    }

    if Y_AXIS > 360 || Y_AXIS == 0 || Y_AXIS < -360{
        Y_AXIS = 0.0
    }else{
        Y_AXIS = Y_AXIS + 0.3*float32(DIRECTION)
    }

    if Z_AXIS > 360 || Z_AXIS == 0 || Z_AXIS < -360{
        Z_AXIS = 0.0
    }else{
        Z_AXIS = Z_AXIS + 0.3 * float32(DIRECTION)
    }
    sdl.GL_SwapBuffers()
    fmt.Println("animate tick")
}

func handleKeyPress(keysym sdl.Keysym) {
    switch keysym.Sym {
    case sdl.K_ESCAPE:
        Quit(0)
    case sdl.K_F1:
        sdl.WM_ToggleFullScreen(surface)
    case sdl.K_a:
        DIRECTION = -1
        Y_AXIS = Y_AXIS - 0.30
    case sdl.K_s:
        DIRECTION = 1
        X_AXIS = X_AXIS + 0.30
    case sdl.K_d:
        DIRECTION = 1
        Y_AXIS = Y_AXIS + 0.30
    case sdl.K_w:
        DIRECTION = -1
        X_AXIS = X_AXIS - 0.30
    case sdl.K_q:
        DIRECTION = 0
    }
}


func initGL() {
    gl.ShadeModel(gl.SMOOTH)
    gl.ClearColor(0.0, 0.0, 0.0, 0.0)
    gl.ClearDepth(1.0)
    gl.Enable(gl.DEPTH_TEST)
    gl.DepthFunc(gl.LESS)
    gl.Hint(gl.PERSPECTIVE_CORRECTION_HINT, gl.NICEST)
}

func create_universe(){
    r := rand.New(rand.NewSource(99))
    for x := range make([]int, SIZE){
        for y := range make([]int, SIZE){
            for z := range make([]int, SIZE){
                if r.Intn(2) == 1{
                    universe[x][y][z] = 1
                }else{
                    universe[x][y][z] = 0
                }
            }
        }
    }
}

func bruteforce_neighbors(i, j, k int)(int){
    neighbors:=0
    for x:=i-1; x<i+2; x++{
        for y:=j-1; y<j+2; y++{
            for z:=k-1; z<k+2; z++{
                if x==i && y==j && z==k{
                    continue
                }
                if x > 0 && x < SIZE && y > 0 && y < SIZE && z > 0 && z < SIZE{
                    if universe[x][y][z] == 1{
                        neighbors += 1
                    }
                }
            }
        }
    }
    return neighbors
}

//this can be optimized using lookup tables and some clever heuristics
func simulate(){
    temp := universe
    var chronon[SIZE][SIZE][SIZE] uint8
    for{
        for x := range make([]int, SIZE){
            for y := range make([]int, SIZE){
                for z := range make([]int, SIZE){
                    chronon[x][y][z] = 0
                    cell := temp[x][y][z]
                    neighborhood:=bruteforce_neighbors(x,y,z)
                    if cell == 1 && (neighborhood < 4 || neighborhood > 5){
                        chronon[x][y][z] = 0
                        continue
                    }
                    if cell == 0 && (neighborhood == 4 || neighborhood == 5){
                        chronon[x][y][z] = 1
                        continue
                    }
                }
            }
        }
        temp = chronon
        mu.Lock()
        universe = temp
        mu.Unlock()
        time.Sleep(500 * time.Millisecond)
        fmt.Println("universe tick")
    }
}


func main(){

    //use ALL the cores!
    runtime.GOMAXPROCS(2)
    // Initialize SDL
    if sdl.Init(sdl.INIT_VIDEO) < 0 {
        panic("Video initialization failed: " + sdl.GetError())
    }
    //if you wish to make an apple pie from scratch, you must first invent the universe
    fmt.Println("create universe")
    create_universe()
    // flags to pass to sdl.SetVideoMode
    videoFlags := sdl.OPENGL    // Enable OpenGL in SDL
    videoFlags |= sdl.DOUBLEBUF // Enable double buffering
    videoFlags |= sdl.HWPALETTE // Store the palette in hardware
    videoFlags |= sdl.RESIZABLE // Enable window resizing

    // get a SDL surface
    surface = sdl.SetVideoMode(SCREEN_WIDTH, SCREEN_HEIGHT, SCREEN_BPP, uint32(videoFlags))

    // verify there is a surface
    if surface == nil {
        panic("Video mode set failed: " + sdl.GetError())
        Quit(1)
    }

    defer Quit(0)
    sdl.GL_SetAttribute(sdl.GL_DOUBLEBUFFER, 1)
    initGL()
    resizeWindow(SCREEN_WIDTH, SCREEN_HEIGHT)

    //simulate universe continuously
    fmt.Println("begin simulation")
    go simulate()
    running := true
    isActive := true
    for running {
        for ev := sdl.PollEvent(); ev != nil; ev = sdl.PollEvent() {
            switch e := ev.(type) {
            case *sdl.ActiveEvent:
                isActive = e.Gain != 0
            case *sdl.ResizeEvent:
                width, height := int(e.W), int(e.H)
                surface = sdl.SetVideoMode(width, height, SCREEN_BPP, uint32(videoFlags))

                if surface == nil {
                    fmt.Println("Could not get a surface after resize:", sdl.GetError())
                    Quit(1)
                }
                resizeWindow(width, height)
            case *sdl.KeyboardEvent:
                if e.Type == sdl.KEYDOWN {
                    handleKeyPress(e.Keysym)
                }
            case *sdl.QuitEvent:
                running = false
            }
        }

        // draw the scene
        if isActive {
            drawGLScene()
        }
    }
}
