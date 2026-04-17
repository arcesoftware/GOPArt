package main

import (
	"log"
	"math"
	"math/rand"
	"runtime"
	"sync"
	"time"
	"unsafe"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

const (
	nParticles = 2500
	numWorkers = 8
	// N-Body Constants
	G         = 400.0 // Gravitational constant
	softening = 25.0  // Prevents "slingshot" errors when particles get too close
)

type Particle struct {
	Pos mgl32.Vec3
	Vel mgl32.Vec3
	Col mgl32.Vec3
}

var (
	window *glfw.Window
	winW   float32 = 1280
	winH   float32 = 720

	particles []Particle
	prog      uint32
	quadProg  uint32

	cubeVAO, cubeVBO uint32
	instanceVBO      uint32
	quadVAO, quadVBO uint32

	sceneFBO, sceneTex, depthRB uint32

	camPos     = mgl32.Vec3{0, 0, 1200}
	yaw, pitch float64 = -90, 0
	speed      float32 = 800
	sens       float64 = 0.15
	cubeSize   float32 = 2.5

	lastX, lastY float64
	firstMouse   = true
	mouseDown    = false
	randomGen    *rand.Rand
)

func init() {
	runtime.LockOSThread()
}

func main() {
	randomGen = rand.New(rand.NewSource(time.Now().UnixNano()))

	if err := glfw.Init(); err != nil {
		log.Fatal(err)
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, gl.TRUE)

	var err error
	window, err = glfw.CreateWindow(int(winW), int(winH), "N-Body Gravitational Cluster", nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()

	if err := gl.Init(); err != nil {
		panic(err)
	}

	gl.Enable(gl.DEPTH_TEST)
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	prog = mustProgram(cubeVS, cubeFS)
	quadProg = mustProgram(quadVS, quadFS)

	initParticles()
	setupCubeGeometry()
	setupQuad()
	setupFBO(int(winW), int(winH))

	window.SetFramebufferSizeCallback(framebufferSizeCallback)
	window.SetCursorPosCallback(mouseMove)
	window.SetMouseButtonCallback(mouseButton)
	window.SetScrollCallback(scrollCallback)

	prev := time.Now()

	for !window.ShouldClose() {
		now := time.Now()
		dt := float32(now.Sub(prev).Seconds())
		prev = now

		// Cap dt to prevent physics explosion during lag
		if dt > 0.033 { dt = 0.033 }

		processInput(dt)
		update(float64(dt))

		renderScene()
		renderComposite()

		window.SwapBuffers()
		glfw.PollEvents()
	}
}

func FastInverseSqrt(x float32) float32 {
	x2, y := x*0.5, x
	i := *(*int32)(unsafe.Pointer(&y))
	i = 0x5f3759df - (i >> 1)
	y = *(*float32)(unsafe.Pointer(&i))
	y = y * (1.5 - (x2 * y * y))
	return y
}

func initParticles() {
	particles = make([]Particle, nParticles)
	radius := float32(400.0)
	for i := range particles {
		a, b := randomGen.Float64()*2*math.Pi, randomGen.Float64()*math.Pi
		particles[i].Pos = mgl32.Vec3{
			radius * float32(math.Cos(a)*math.Sin(b)),
			radius * float32(math.Sin(a)*math.Sin(b)),
			radius * float32(math.Cos(b)),
		}
		// Initial orbital velocity (slight swirl)
		particles[i].Vel = particles[i].Pos.Cross(mgl32.Vec3{0, 1, 0}).Normalize().Mul(50.0)
		particles[i].Col = mgl32.Vec3{0.4, 0.7, 1.0}
	}
}

func update(dt float64) {
	var wg sync.WaitGroup
	chunk := len(particles) / numWorkers
	fdt := float32(dt)

	for w := 0; w < numWorkers; w++ {
		start, end := w*chunk, (w+1)*chunk
		if w == numWorkers-1 { end = len(particles) }
		wg.Add(1)
		go func(s, e int) {
			defer wg.Done()
			for i := s; i < e; i++ {
				p := &particles[i]
				var totalAcc mgl32.Vec3

				// N-Body Calculation: Each particle pulls on p
				for j := 0; j < nParticles; j++ {
					if i == j { continue }
					
					diff := particles[j].Pos.Sub(p.Pos)
					distSq := diff.Dot(diff) + softening
					
					// Acceleration = G * m * (diff / dist^3)
					// Using Quake trick: invDist = 1/sqrt(distSq)
					invDist := FastInverseSqrt(distSq)
					invDistCube := invDist * invDist * invDist
					
					acc := diff.Mul(G * invDistCube)
					totalAcc = totalAcc.Add(acc)
				}

				p.Vel = p.Vel.Add(totalAcc.Mul(fdt))
				p.Pos = p.Pos.Add(p.Vel.Mul(fdt))
				
				// System damping (friction) to keep orbits stable
				p.Vel = p.Vel.Mul(0.99)

				// Color by velocity magnitude
				vLen := p.Vel.Len()
				p.Col = mgl32.Vec3{
					0.3 + vLen*0.005,
					0.5 + vLen*0.002,
					0.9 + (p.Pos.Z()/800.0),
				}
			}
		}(start, end)
	}
	wg.Wait()
}

func processInput(dt float32) {
	front := getCameraFront()
	right := front.Cross(mgl32.Vec3{0, 1, 0}).Normalize()
	actualSpeed := speed * dt
	if window.GetKey(glfw.KeyW) == glfw.Press { camPos = camPos.Add(front.Mul(actualSpeed)) }
	if window.GetKey(glfw.KeyS) == glfw.Press { camPos = camPos.Sub(front.Mul(actualSpeed)) }
	if window.GetKey(glfw.KeyA) == glfw.Press { camPos = camPos.Sub(right.Mul(actualSpeed)) }
	if window.GetKey(glfw.KeyD) == glfw.Press { camPos = camPos.Add(right.Mul(actualSpeed)) }
	
	if window.GetKey(glfw.KeyP) == glfw.Press { cubeSize += 0.1 }
	if window.GetKey(glfw.KeyO) == glfw.Press { cubeSize -= 0.1; if cubeSize < 0.1 { cubeSize = 0.1 } }
	
	// RESET KEY
	if window.GetKey(glfw.KeyR) == glfw.Press { initParticles() }
}

func getCameraFront() mgl32.Vec3 {
	radYaw, radPitch := mgl32.DegToRad(float32(yaw)), mgl32.DegToRad(float32(pitch))
	return mgl32.Vec3{
		float32(math.Cos(float64(radYaw)) * math.Cos(float64(radPitch))),
		float32(math.Sin(float64(radPitch))),
		float32(math.Sin(float64(radYaw)) * math.Cos(float64(radPitch))),
	}.Normalize()
}

func renderScene() {
	gl.BindFramebuffer(gl.FRAMEBUFFER, sceneFBO)
	gl.Viewport(0, 0, int32(winW), int32(winH))
	gl.ClearColor(0, 0, 0, 1)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	gl.UseProgram(prog)
	proj := mgl32.Perspective(mgl32.DegToRad(45), winW/winH, 1.0, 10000.0)
	view := camera()
	
	gl.UniformMatrix4fv(gl.GetUniformLocation(prog, gl.Str("uProj\x00")), 1, false, &proj[0])
	gl.UniformMatrix4fv(gl.GetUniformLocation(prog, gl.Str("uView\x00")), 1, false, &view[0])
	gl.Uniform1f(gl.GetUniformLocation(prog, gl.Str("uSize\x00")), cubeSize)

	data := make([]float32, 0, nParticles*6)
	for _, p := range particles {
		data = append(data, p.Pos.X(), p.Pos.Y(), p.Pos.Z(), p.Col.X(), p.Col.Y(), p.Col.Z())
	}
	gl.BindBuffer(gl.ARRAY_BUFFER, instanceVBO)
	gl.BufferSubData(gl.ARRAY_BUFFER, 0, len(data)*4, gl.Ptr(data))

	gl.BindVertexArray(cubeVAO)
	gl.DrawArraysInstanced(gl.TRIANGLES, 0, 36, int32(nParticles))
}

func setupCubeGeometry() {
	vertices := []float32{
		-0.5, -0.5, -0.5,  0,  0, -1,  0.5, -0.5, -0.5,  0,  0, -1,  0.5,  0.5, -0.5,  0,  0, -1,
		 0.5,  0.5, -0.5,  0,  0, -1, -0.5,  0.5, -0.5,  0,  0, -1, -0.5, -0.5, -0.5,  0,  0, -1,
		-0.5, -0.5,  0.5,  0,  0,  1,  0.5, -0.5,  0.5,  0,  0,  1,  0.5,  0.5,  0.5,  0,  0,  1,
		 0.5,  0.5,  0.5,  0,  0,  1, -0.5,  0.5,  0.5,  0,  0,  1, -0.5, -0.5,  0.5,  0,  0,  1,
		-0.5,  0.5,  0.5, -1,  0,  0, -0.5,  0.5, -0.5, -1,  0,  0, -0.5, -0.5, -0.5, -1,  0,  0,
		-0.5, -0.5, -0.5, -1,  0,  0, -0.5, -0.5,  0.5, -1,  0,  0, -0.5,  0.5,  0.5, -1,  0,  0,
		 0.5,  0.5,  0.5,  1,  0,  0,  0.5,  0.5, -0.5,  1,  0,  0,  0.5, -0.5, -0.5,  1,  0,  0,
		 0.5, -0.5, -0.5,  1,  0,  0,  0.5, -0.5,  0.5,  1,  0,  0,  0.5,  0.5,  0.5,  1,  0,  0,
		-0.5, -0.5, -0.5,  0, -1,  0,  0.5, -0.5, -0.5,  0, -1,  0,  0.5, -0.5,  0.5,  0, -1,  0,
		 0.5, -0.5,  0.5,  0, -1,  0, -0.5, -0.5,  0.5,  0, -1,  0, -0.5, -0.5, -0.5,  0, -1,  0,
		-0.5,  0.5, -0.5,  0,  1,  0,  0.5,  0.5, -0.5,  0,  1,  0,  0.5,  0.5,  0.5,  0,  1,  0,
		 0.5,  0.5,  0.5,  0,  1,  0, -0.5,  0.5,  0.5,  0,  1,  0, -0.5,  0.5, -0.5,  0,  1,  0,
	}
	gl.GenVertexArrays(1, &cubeVAO); gl.BindVertexArray(cubeVAO)
	gl.GenBuffers(1, &cubeVBO); gl.BindBuffer(gl.ARRAY_BUFFER, cubeVBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)
	gl.EnableVertexAttribArray(0); gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 6*4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(1); gl.VertexAttribPointer(1, 3, gl.FLOAT, false, 6*4, gl.PtrOffset(3*4))
	gl.GenBuffers(1, &instanceVBO); gl.BindBuffer(gl.ARRAY_BUFFER, instanceVBO)
	gl.BufferData(gl.ARRAY_BUFFER, nParticles*6*4, nil, gl.DYNAMIC_DRAW)
	gl.EnableVertexAttribArray(2); gl.VertexAttribPointer(2, 3, gl.FLOAT, false, 6*4, gl.PtrOffset(0))
	gl.VertexAttribDivisor(2, 1)
	gl.EnableVertexAttribArray(3); gl.VertexAttribPointer(3, 3, gl.FLOAT, false, 6*4, gl.PtrOffset(3*4))
	gl.VertexAttribDivisor(3, 1)
}

func camera() mgl32.Mat4 { return mgl32.LookAtV(camPos, camPos.Add(getCameraFront()), mgl32.Vec3{0, 1, 0}) }
func renderComposite() { gl.BindFramebuffer(gl.FRAMEBUFFER, 0); gl.Clear(gl.COLOR_BUFFER_BIT); gl.Disable(gl.DEPTH_TEST); gl.UseProgram(quadProg); gl.ActiveTexture(gl.TEXTURE0); gl.BindTexture(gl.TEXTURE_2D, sceneTex); gl.BindVertexArray(quadVAO); gl.DrawArrays(gl.TRIANGLES, 0, 6); gl.Enable(gl.DEPTH_TEST) }
func setupQuad() { q := []float32{-1,-1,0,0, 1,-1,1,0, 1,1,1,1, -1,-1,0,0, 1,1,1,1, -1,1,0,1}; gl.GenVertexArrays(1, &quadVAO); gl.BindVertexArray(quadVAO); gl.GenBuffers(1, &quadVBO); gl.BindBuffer(gl.ARRAY_BUFFER, quadVBO); gl.BufferData(gl.ARRAY_BUFFER, len(q)*4, gl.Ptr(q), gl.STATIC_DRAW); gl.EnableVertexAttribArray(0); gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 4*4, gl.PtrOffset(0)); gl.EnableVertexAttribArray(1); gl.VertexAttribPointer(1, 2, gl.FLOAT, false, 4*4, gl.PtrOffset(2*4)) }
func setupFBO(w, h int) { gl.GenFramebuffers(1, &sceneFBO); gl.BindFramebuffer(gl.FRAMEBUFFER, sceneFBO); gl.GenTextures(1, &sceneTex); gl.BindTexture(gl.TEXTURE_2D, sceneTex); gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA16F, int32(w), int32(h), 0, gl.RGBA, gl.FLOAT, nil); gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR); gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, sceneTex, 0); gl.GenRenderbuffers(1, &depthRB); gl.BindRenderbuffer(gl.RENDERBUFFER, depthRB); gl.RenderbufferStorage(gl.RENDERBUFFER, gl.DEPTH_COMPONENT, int32(w), int32(h)); gl.FramebufferRenderbuffer(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, gl.RENDERBUFFER, depthRB); gl.BindFramebuffer(gl.FRAMEBUFFER, 0) }
func resizeFBO(w, h int) { gl.DeleteTextures(1, &sceneTex); gl.DeleteRenderbuffers(1, &depthRB); gl.DeleteFramebuffers(1, &sceneFBO); setupFBO(w, h) }
func mouseMove(w *glfw.Window, x, y float64) { if !mouseDown { firstMouse = true; return }; if firstMouse { lastX, lastY = x, y; firstMouse = false }; dx, dy := (x-lastX)*sens, (lastY-y)*sens; lastX, lastY = x, y; yaw, pitch = yaw+dx, pitch+dy; if pitch > 89 { pitch = 89 }; if pitch < -89 { pitch = -89 } }
func mouseButton(w *glfw.Window, b glfw.MouseButton, a glfw.Action, m glfw.ModifierKey) { mouseDown = (b == glfw.MouseButtonLeft && a == glfw.Press) }
func scrollCallback(w *glfw.Window, x, y float64) { speed += float32(y * 50); if speed < 50 { speed = 50 } }
func framebufferSizeCallback(w *glfw.Window, width, height int) { if width == 0 || height == 0 { return }; winW, winH = float32(width), float32(height); gl.Viewport(0, 0, int32(width), int32(height)); resizeFBO(width, height) }
func mustProgram(vs, fs string) uint32 { v, f := compile(vs, gl.VERTEX_SHADER), compile(fs, gl.FRAGMENT_SHADER); p := gl.CreateProgram(); gl.AttachShader(p, v); gl.AttachShader(p, f); gl.LinkProgram(p); return p }
func compile(src string, t uint32) uint32 { s := gl.CreateShader(t); cs, free := gl.Strs(src); gl.ShaderSource(s, 1, cs, nil); free(); gl.CompileShader(s); return s }

var cubeVS = `#version 410 core
layout(location=0) in vec3 inPos;
layout(location=1) in vec3 inNormal;
layout(location=2) in vec3 instPos;
layout(location=3) in vec3 instCol;
uniform mat4 uProj, uView;
uniform float uSize;
out vec3 vCol, vNormal;
void main(){
    vNormal = inNormal;
    vCol = instCol;
    gl_Position = uProj * uView * vec4(instPos + (inPos * uSize), 1.0);
}` + "\x00"

var cubeFS = `#version 410 core
in vec3 vCol, vNormal;
out vec4 fragColor;
void main(){
    vec3 light = normalize(vec3(0.5, 1.0, 0.5));
    float d = max(dot(vNormal, light), 0.2);
    fragColor = vec4(vCol * d, 1.0);
}` + "\x00"

var quadVS = `#version 410 core
layout(location=0) in vec2 pos;
layout(location=1) in vec2 uvIn;
out vec2 uv;
void main(){ uv = uvIn; gl_Position = vec4(pos, 0, 1); }` + "\x00"

var quadFS = `#version 410 core
in vec2 uv;
out vec4 fragColor;
uniform sampler2D uTex;
void main(){ fragColor = vec4(texture(uTex, uv).rgb * 1.5, 1.0); }` + "\x00"