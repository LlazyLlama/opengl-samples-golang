package main

/*
Adapted from this tutorial: http://www.learnopengl.com/#!Getting-started/Camera

Shows how to create a basic controllable FPS camera
*/

import (
	"log"
	"runtime"
	"math"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.1/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"

	"github.com/opengl-samples-golang/basic-camera/gfx"
)

const windowWidth int  = 1280
const windowHeight int = 720

// only using global variables because this is meant as a simple example
var cameraPos   = mgl32.Vec3{0.0, 0.0, 3.0}
var cameraFront = mgl32.Vec3{}
var cameraUp    = mgl32.Vec3{0.0, 1.0, 0.0}

var keysPressed [glfw.KeyLast]bool

var firstCursorAction = true
var cursorLastX float64
var cursorLastY float64
var pitch float64 = 0
var yaw float64 = -90

// vertices to draw 6 faces of a cube
var cubeVertices = []float32{
	// position        // texture position
	-0.5, -0.5, -0.5,  0.0, 0.0,
	 0.5, -0.5, -0.5,  1.0, 0.0,
	 0.5,  0.5, -0.5,  1.0, 1.0,
	 0.5,  0.5, -0.5,  1.0, 1.0,
	-0.5,  0.5, -0.5,  0.0, 1.0,
	-0.5, -0.5, -0.5,  0.0, 0.0,

	-0.5, -0.5,  0.5,  0.0, 0.0,
	 0.5, -0.5,  0.5,  1.0, 0.0,
	 0.5,  0.5,  0.5,  1.0, 1.0,
	 0.5,  0.5,  0.5,  1.0, 1.0,
	-0.5,  0.5,  0.5,  0.0, 1.0,
	-0.5, -0.5,  0.5,  0.0, 0.0,

	-0.5,  0.5,  0.5,  1.0, 0.0,
	-0.5,  0.5, -0.5,  1.0, 1.0,
	-0.5, -0.5, -0.5,  0.0, 1.0,
	-0.5, -0.5, -0.5,  0.0, 1.0,
	-0.5, -0.5,  0.5,  0.0, 0.0,
	-0.5,  0.5,  0.5,  1.0, 0.0,

	 0.5,  0.5,  0.5,  1.0, 0.0,
	 0.5,  0.5, -0.5,  1.0, 1.0,
	 0.5, -0.5, -0.5,  0.0, 1.0,
	 0.5, -0.5, -0.5,  0.0, 1.0,
	 0.5, -0.5,  0.5,  0.0, 0.0,
	 0.5,  0.5,  0.5,  1.0, 0.0,

	-0.5, -0.5, -0.5,  0.0, 1.0,
	 0.5, -0.5, -0.5,  1.0, 1.0,
	 0.5, -0.5,  0.5,  1.0, 0.0,
	 0.5, -0.5,  0.5,  1.0, 0.0,
	-0.5, -0.5,  0.5,  0.0, 0.0,
	-0.5, -0.5, -0.5,  0.0, 1.0,

	-0.5,  0.5, -0.5,  0.0, 1.0,
	 0.5,  0.5, -0.5,  1.0, 1.0,
	 0.5,  0.5,  0.5,  1.0, 0.0,
	 0.5,  0.5,  0.5,  1.0, 0.0,
	-0.5,  0.5,  0.5,  0.0, 0.0,
	-0.5,  0.5, -0.5,  0.0, 1.0,
}

var cubePositions = [][]float32 {
	{ 0.0,  0.0,  -3.0},
	{ 2.0,  5.0, -15.0},
	{-1.5, -2.2, -2.5 },
	{-3.8, -2.0, -12.3},
	{ 2.4, -0.4, -3.5 },
	{-1.7,  3.0, -7.5 },
	{ 1.3, -2.0, -2.5 },
	{ 1.5,  2.0, -2.5 },
	{ 1.5,  0.2, -1.5 },
	{-1.3,  1.0, -1.5 },
}

func init() {
	// GLFW event handling must be run on the main OS thread
	runtime.LockOSThread()
}

func main() {
	if err := glfw.Init(); err != nil {
		log.Fatalln("failed to inifitialize glfw:", err)
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	window, err := glfw.CreateWindow(windowWidth, windowHeight, "basic camera", nil, nil)
	if err != nil {
		log.Fatalln(err)
	}

	window.MakeContextCurrent()
	window.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
	window.SetKeyCallback(keyCallback)
	window.SetCursorPosCallback(mouseCallback)

	// Initialize Glow (go function bindings)
	if err := gl.Init(); err != nil {
		panic(err)
	}

	err = programLoop(window)
	if err != nil {
		log.Fatalln(err)
	}
}

/*
 * Creates the Vertex Array Object for a triangle.
 * indices is leftover from earlier samples and not used here.
 */
func createVAO(vertices []float32, indices []uint32) uint32 {

	var VAO uint32
	gl.GenVertexArrays(1, &VAO)

	var VBO uint32
	gl.GenBuffers(1, &VBO)

	var EBO uint32;
	gl.GenBuffers(1, &EBO)

	// Bind the Vertex Array Object first, then bind and set vertex buffer(s) and attribute pointers()
	gl.BindVertexArray(VAO)

	// copy vertices data into VBO (it needs to be bound first)
	gl.BindBuffer(gl.ARRAY_BUFFER, VBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	// size of one whole vertex (sum of attrib sizes)
	var stride int32 = 3*4 + 2*4
	var offset int = 0

	// position
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, stride, gl.PtrOffset(offset))
	gl.EnableVertexAttribArray(0)
	offset += 3*4

	// texture position
	gl.VertexAttribPointer(1, 2, gl.FLOAT, false, stride, gl.PtrOffset(offset))
	gl.EnableVertexAttribArray(1)
	offset += 2*4

	// unbind the VAO (safe practice so we don't accidentally (mis)configure it later)
	gl.BindVertexArray(0)

	return VAO
}

func programLoop(window *glfw.Window) error {

	// the linked shader program determines how the data will be rendered
	vertShader, err := gfx.NewShaderFromFile("shaders/basic.vert", gl.VERTEX_SHADER)
	if err != nil {
		return err
	}

	fragShader, err := gfx.NewShaderFromFile("shaders/basic.frag", gl.FRAGMENT_SHADER)
	if err != nil {
		return err
	}

	program, err := gfx.NewProgram(vertShader, fragShader)
	if err != nil {
		return err
	}
	defer program.Delete()

	VAO := createVAO(cubeVertices, nil)
	texture0, err := gfx.NewTextureFromFile("../images/RTS_Crate.png",
	                                        gl.CLAMP_TO_EDGE, gl.CLAMP_TO_EDGE)
	if err != nil {
		panic(err.Error())
	}

	texture1, err := gfx.NewTextureFromFile("../images/trollface-transparent.png",
	                                        gl.CLAMP_TO_EDGE, gl.CLAMP_TO_EDGE)
	if err != nil {
		panic(err.Error())
	}

	// ensure that triangles that are "behind" others do not draw over top of them
	gl.Enable(gl.DEPTH_TEST)

	lastFrameTime := glfw.GetTime()

	for !window.ShouldClose() {
		// poll events and call their registered callbacks
		glfw.PollEvents()

		// base calculations of time since last frame (basic program loop idea)
		// For better advanced impl, read: http://gafferongames.com/game-physics/fix-your-timestep/
		curFrameTime  := glfw.GetTime()
		dTime         := curFrameTime - lastFrameTime
		lastFrameTime  = curFrameTime

		// update global variables about camera target and position
		updateCamera()
		updateMovement(dTime)

		// background color
		gl.ClearColor(0.2, 0.5, 0.5, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)  // depth buffer needed for DEPTH_TEST

		program.Use()

		// bind textures
		texture0.Bind(gl.TEXTURE0)
		texture0.SetUniform(program.GetUniformLocation("ourTexture0"))

		texture1.Bind(gl.TEXTURE1)
		texture1.SetUniform(program.GetUniformLocation("ourTexture1"))

		// cube rotation matrices
		rotateX   := (mgl32.Rotate3DX(mgl32.DegToRad(-60 * float32(glfw.GetTime()))))
		rotateY   := (mgl32.Rotate3DY(mgl32.DegToRad(-60 * float32(glfw.GetTime()))))
		rotateZ   := (mgl32.Rotate3DZ(mgl32.DegToRad(-60 * float32(glfw.GetTime()))))

		// creates perspective
		fov := float32(60.0)
		projectTransform := mgl32.Perspective(mgl32.DegToRad(fov),
		                                      float32(windowWidth)/float32(windowHeight),
		                                      0.1,
		                                      100.0)

		cameraTarget := cameraPos.Add(cameraFront)  // TODO-cs: why?

		cameraTransform := mgl32.LookAt(
			cameraPos.X(), cameraPos.Y(), cameraPos.Z(),
			cameraTarget.X(), cameraTarget.Y(), cameraTarget.Z(),
			cameraUp.X(), cameraUp.Y(), cameraUp.Z(),
		)

		gl.UniformMatrix4fv(program.GetUniformLocation("camera"), 1, false, &cameraTransform[0])
		gl.UniformMatrix4fv(program.GetUniformLocation("project"), 1, false,
		&projectTransform[0])

		gl.BindVertexArray(VAO)

		// draw each cube after all coordinate system transforms are bound
		for _, pos := range cubePositions {
			worldTranslate := mgl32.Translate3D(pos[0], pos[1], pos[2])
			worldTransform := (worldTranslate.Mul4(rotateX.Mul3(rotateY).Mul3(rotateZ).Mat4()))

			gl.UniformMatrix4fv(program.GetUniformLocation("world"), 1, false,
			                    &worldTransform[0])

			gl.DrawArrays(gl.TRIANGLES, 0, 36)
		}

		gl.BindVertexArray(0)

		texture0.UnBind()
		texture1.UnBind()

		// end of draw loop

		// swap in the rendered buffer
		window.SwapBuffers()
	}

	return nil
}

func updateCamera() {
	// x, y, z
	cameraFront[0] = float32(math.Cos(mgl64.DegToRad(pitch)) * math.Cos(mgl64.DegToRad(yaw)))
	cameraFront[1] = float32(math.Sin(mgl64.DegToRad(pitch)))
	cameraFront[2] = float32(math.Cos(mgl64.DegToRad(pitch)) * math.Sin(mgl64.DegToRad(yaw)))
	cameraFront = cameraFront.Normalize()
}

func updateMovement(dTime float64) {
	cameraSpeed := 5.00
	adjustedSpeed := float32(dTime * cameraSpeed)

	if keysPressed[glfw.KeyW] {
		cameraPos = cameraPos.Add(cameraFront.Mul(adjustedSpeed))
	}
	if keysPressed[glfw.KeyS] {
		cameraPos = cameraPos.Sub(cameraFront.Mul(adjustedSpeed))
	}
	if keysPressed[glfw.KeyA] {
		cameraPos = cameraPos.Sub(cameraFront.Cross(cameraUp).Normalize().Mul(adjustedSpeed))
	}
	if keysPressed[glfw.KeyD] {
		cameraPos = cameraPos.Add(cameraFront.Cross(cameraUp).Normalize().Mul(adjustedSpeed))
	}
}

func mouseCallback(window *glfw.Window, xpos, ypos float64) {
	sensitivity := 0.05

	if firstCursorAction {
		cursorLastX = xpos
		cursorLastY = ypos
		firstCursorAction = false
	}

	dx := sensitivity * (xpos - cursorLastX)
	dy := sensitivity * -(ypos - cursorLastY) // reversed since y goes from bottom to top

	cursorLastX = xpos
	cursorLastY = ypos

	pitch += dy
	if pitch > 89.0 {
		pitch = 89.0
	} else if pitch < -89.0 {
		pitch = -89.0
	}

	yaw = math.Mod(yaw + dx, 360)
}

func keyCallback(window *glfw.Window, key glfw.Key, scancode int, action glfw.Action,
	mods glfw.ModifierKey) {

	// When a user presses the escape key, we set the WindowShouldClose property to true,
	// which closes the application
	if key == glfw.KeyEscape && action == glfw.Press {
		window.SetShouldClose(true)
	}

	// timing for key events occurs differently from what the program loop requires
	// so just track what key actions occur and then access them in the program loop
	switch action {
	case glfw.Press:
		keysPressed[key] = true
	case glfw.Release:
		keysPressed[key] = false
	}
}
