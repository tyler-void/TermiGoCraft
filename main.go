package main

import (
	"log"
	"math"
	"time"

	"github.com/nsf/termbox-go"
)

// Vector3 represents a point in 3D space
type Vector3 struct {
	X, Y, Z float64
}

// Add returns the sum of two vectors
func (v Vector3) Add(other Vector3) Vector3 {
	return Vector3{
		X: v.X + other.X,
		Y: v.Y + other.Y,
		Z: v.Z + other.Z,
	}
}

// Scale returns the vector scaled by a factor
func (v Vector3) Scale(factor float64) Vector3 {
	return Vector3{
		X: v.X * factor,
		Y: v.Y * factor,
		Z: v.Z * factor,
	}
}

// Normalize returns a unit vector in the same direction
func (v Vector3) Normalize() Vector3 {
	length := math.Sqrt(v.X*v.X + v.Y*v.Y + v.Z*v.Z)
	if length == 0 {
		return Vector3{0, 0, 0}
	}
	return Vector3{
		X: v.X / length,
		Y: v.Y / length,
		Z: v.Z / length,
	}
}

// Cross returns the cross product of two vectors
func (v Vector3) Cross(other Vector3) Vector3 {
	return Vector3{
		X: v.Y*other.Z - v.Z*other.Y,
		Y: v.Z*other.X - v.X*other.Z,
		Z: v.X*other.Y - v.Y*other.X,
	}
}

// Vertex represents a 3D vertex with position
type Vertex struct {
	Position Vector3
	Color    termbox.Attribute
}

// Triangle represents a 3D triangle
type Triangle struct {
	Vertices [3]Vertex
}

// Camera represents the viewer's position and orientation
type Camera struct {
	Position Vector3
	Yaw      float64 // Horizontal rotation in radians
	Pitch    float64 // Vertical rotation in radians
	FOV      float64
	NearClip float64
	FarClip  float64
}

// GetDirection returns the camera's forward vector
func (c Camera) GetDirection() Vector3 {
	return Vector3{
		X: math.Cos(c.Pitch) * math.Sin(c.Yaw),
		Y: math.Sin(c.Pitch),
		Z: math.Cos(c.Pitch) * math.Cos(c.Yaw),
	}
}

// GetRight returns the camera's right vector
func (c Camera) GetRight() Vector3 {
	// The right vector is perpendicular to both the direction and world-up
	worldUp := Vector3{0, 1, 0}
	forward := c.GetDirection()

	// Cross product of forward and world-up gives the right vector
	return forward.Cross(worldUp).Normalize()
}

// GetUp returns the camera's up vector
func (c Camera) GetUp() Vector3 {
	// The camera's up vector is perpendicular to both forward and right
	forward := c.GetDirection()
	right := c.GetRight()

	// Cross product of right and forward gives the up vector
	return right.Cross(forward).Normalize()
}

// GameObject represents an object in the 3D world
type GameObject struct {
	Position Vector3
	Mesh     []Triangle
}

// World holds all game objects
type World struct {
	Camera  Camera
	Player  GameObject
	Objects []GameObject
	ZBuffer [][]float64 // For depth comparison
}

func main() {
	err := termbox.Init()
	if err != nil {
		log.Fatalf("Cannot initialize termbox: %v", err)
	}
	defer termbox.Close()

	termbox.SetInputMode(termbox.InputEsc)

	// Initialize the world
	world := initializeWorld()

	// Create a Z-buffer
	width, height := termbox.Size()
	zBuffer := make([][]float64, width)
	for i := range zBuffer {
		zBuffer[i] = make([]float64, height)
		for j := range zBuffer[i] {
			zBuffer[i][j] = math.Inf(1) // Initialize to infinity
		}
	}
	world.ZBuffer = zBuffer

	// Draw the initial state
	render(world)

	// Event handling
	eventQueue := make(chan termbox.Event)
	go func() {
		for {
			eventQueue <- termbox.PollEvent()
		}
	}()

	// Game loop
	running := true
	// lastTick := time.Now()

	for running {
		// now := time.Now()
		// deltaTime := now.Sub(lastTick).Seconds()
		// lastTick = now

		select {
		case ev := <-eventQueue:
			if ev.Type == termbox.EventKey {
				switch {
				case ev.Key == termbox.KeyEsc:
					running = false

				// Camera rotation with arrow keys
				case ev.Key == termbox.KeyArrowLeft:
					world.Camera.Yaw -= 0.1
				case ev.Key == termbox.KeyArrowRight:
					world.Camera.Yaw += 0.1
				case ev.Key == termbox.KeyArrowUp:
					world.Camera.Pitch += 0.1
					// Cap pitch to avoid gimbal lock
					if world.Camera.Pitch > 1.5 {
						world.Camera.Pitch = 1.5
					}
				case ev.Key == termbox.KeyArrowDown:
					world.Camera.Pitch -= 0.1
					if world.Camera.Pitch < -1.5 {
						world.Camera.Pitch = -1.5
					}

				// Camera movement with WASD + Ctrl
				case ev.Ch == 'w' && ev.Mod&termbox.ModAlt != 0:
					moveCameraForward(&world.Camera, 0.5)
				case ev.Ch == 's' && ev.Mod&termbox.ModAlt != 0:
					moveCameraForward(&world.Camera, -0.5)
				case ev.Ch == 'a' && ev.Mod&termbox.ModAlt != 0:
					moveCameraRight(&world.Camera, -0.5)
				case ev.Ch == 'd' && ev.Mod&termbox.ModAlt != 0:
					moveCameraRight(&world.Camera, 0.5)
				case ev.Key == termbox.KeyPgup:
					world.Camera.Position.Y += 0.5
				case ev.Key == termbox.KeyPgdn:
					world.Camera.Position.Y -= 0.5

				// Player object movement with WASD (no Ctrl)
				case ev.Ch == 'w' && ev.Mod&termbox.ModAlt == 0:
					movePlayerRelativeToCamera(&world.Player, world.Camera, 0, 0, 0.5)
				case ev.Ch == 's' && ev.Mod&termbox.ModAlt == 0:
					movePlayerRelativeToCamera(&world.Player, world.Camera, 0, 0, -0.5)
				case ev.Ch == 'a' && ev.Mod&termbox.ModAlt == 0:
					movePlayerRelativeToCamera(&world.Player, world.Camera, -0.5, 0, 0)
				case ev.Ch == 'd' && ev.Mod&termbox.ModAlt == 0:
					movePlayerRelativeToCamera(&world.Player, world.Camera, 0.5, 0, 0)
				case ev.Ch == 'q' && ev.Mod&termbox.ModAlt == 0:
					world.Player.Position.Y -= 0.5
				case ev.Ch == 'e' && ev.Mod&termbox.ModAlt == 0:
					world.Player.Position.Y += 0.5
				}

				render(world)
			} else if ev.Type == termbox.EventResize {
				// Resize Z-buffer
				width, height = termbox.Size()
				zBuffer = make([][]float64, width)
				for i := range zBuffer {
					zBuffer[i] = make([]float64, height)
					for j := range zBuffer[i] {
						zBuffer[i][j] = math.Inf(1)
					}
				}
				world.ZBuffer = zBuffer
				render(world)
			}
		default:
			// Animation updates could go here
			time.Sleep(16 * time.Millisecond) // ~60 FPS
		}
	}
}

func movePlayerRelativeToCamera(player *GameObject, camera Camera, rightAmount, upAmount, forwardAmount float64) {
	// Get the camera's orientation vectors
	forward := camera.GetDirection()
	right := camera.GetRight()
	up := camera.GetUp()

	// Remove Y component for forward movement to keep it on ground plane
	forward.Y = 0
	forward = forward.Normalize()

	// Calculate movement vector
	movement := Vector3{0, 0, 0}
	movement = movement.Add(right.Scale(rightAmount))
	movement = movement.Add(up.Scale(upAmount))
	movement = movement.Add(forward.Scale(forwardAmount))

	// Apply movement
	player.Position = player.Position.Add(movement)
}

func moveCameraForward(camera *Camera, amount float64) {
	direction := camera.GetDirection()
	camera.Position = camera.Position.Add(direction.Scale(amount))
}

func moveCameraRight(camera *Camera, amount float64) {
	right := camera.GetRight()
	camera.Position = camera.Position.Add(right.Scale(amount))
}

func initializeWorld() World {
	// Create a simple cube mesh for player
	playerMesh := createCubeMesh(termbox.ColorBlue)

	// Create another cube for a scene object
	objectMesh := createCubeMesh(termbox.ColorRed)

	return World{
		Camera: Camera{
			Position: Vector3{0, 1, -5},
			Yaw:      0,
			Pitch:    0,
			FOV:      90.0,
			NearClip: 0.1,
			FarClip:  100.0,
		},
		Player: GameObject{
			Position: Vector3{0, 0, 0},
			Mesh:     playerMesh,
		},
		Objects: []GameObject{
			{
				Position: Vector3{3, 0, 3},
				Mesh:     objectMesh,
			},
			{
				Position: Vector3{-3, 0, 3},
				Mesh:     objectMesh,
			},
			{
				Position: Vector3{0, 0, 6},
				Mesh:     objectMesh,
			},
		},
	}
}

func createCubeMesh(color termbox.Attribute) []Triangle {
	// Define cube vertices
	vertices := []Vertex{
		// Front face
		{Position: Vector3{-0.5, -0.5, 0.5}, Color: color}, // 0
		{Position: Vector3{0.5, -0.5, 0.5}, Color: color},  // 1
		{Position: Vector3{0.5, 0.5, 0.5}, Color: color},   // 2
		{Position: Vector3{-0.5, 0.5, 0.5}, Color: color},  // 3

		// Back face
		{Position: Vector3{-0.5, -0.5, -0.5}, Color: color}, // 4
		{Position: Vector3{0.5, -0.5, -0.5}, Color: color},  // 5
		{Position: Vector3{0.5, 0.5, -0.5}, Color: color},   // 6
		{Position: Vector3{-0.5, 0.5, -0.5}, Color: color},  // 7
	}

	// Define triangles using indices
	triangles := []Triangle{
		// Front face
		{Vertices: [3]Vertex{vertices[0], vertices[1], vertices[2]}},
		{Vertices: [3]Vertex{vertices[0], vertices[2], vertices[3]}},

		// Back face
		{Vertices: [3]Vertex{vertices[5], vertices[4], vertices[7]}},
		{Vertices: [3]Vertex{vertices[5], vertices[7], vertices[6]}},

		// Left face
		{Vertices: [3]Vertex{vertices[4], vertices[0], vertices[3]}},
		{Vertices: [3]Vertex{vertices[4], vertices[3], vertices[7]}},

		// Right face
		{Vertices: [3]Vertex{vertices[1], vertices[5], vertices[6]}},
		{Vertices: [3]Vertex{vertices[1], vertices[6], vertices[2]}},

		// Top face
		{Vertices: [3]Vertex{vertices[3], vertices[2], vertices[6]}},
		{Vertices: [3]Vertex{vertices[3], vertices[6], vertices[7]}},

		// Bottom face
		{Vertices: [3]Vertex{vertices[4], vertices[5], vertices[1]}},
		{Vertices: [3]Vertex{vertices[4], vertices[1], vertices[0]}},
	}

	return triangles
}

func render(world World) {
	// Clear screen and Z-buffer
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	width, height := termbox.Size()
	for i := range world.ZBuffer {
		for j := range world.ZBuffer[i] {
			world.ZBuffer[i][j] = math.Inf(1)
		}
	}

	// Draw a ground grid for orientation
	drawGrid(world.Camera, width, height, world.ZBuffer)

	// Draw player
	drawObject(world.Player, world.Camera, width, height, world.ZBuffer)

	// For each object in the world
	for _, obj := range world.Objects {
		drawObject(obj, world.Camera, width, height, world.ZBuffer)
	}

	// Draw instructions
	drawString(1, height-3, "Arrow keys: Rotate camera | Ctrl+WASD: Move camera | PgUp/PgDn: Camera up/down", termbox.ColorYellow, termbox.ColorDefault)
	drawString(1, height-2, "WASD: Move player | Q/E: Player down/up | ESC: Quit", termbox.ColorYellow, termbox.ColorDefault)

	termbox.Flush()
}

func drawObject(obj GameObject, camera Camera, screenWidth, screenHeight int, zBuffer [][]float64) {
	// For each triangle in the mesh
	for _, triangle := range obj.Mesh {
		// Project and draw the triangle
		drawTriangle(triangle, obj.Position, camera, screenWidth, screenHeight, zBuffer)
	}
}

func drawGrid(camera Camera, screenWidth, screenHeight int, zBuffer [][]float64) {
	gridSize := 10.0 // Changed to float64
	gridSpacing := 1.0
	gridColor := termbox.ColorWhite

	// Create grid lines along X and Z axes
	for i := -gridSize; i <= gridSize; i += 1.0 { // Changed loop to use float64
		// Lines along X axis
		startPos := Vector3{-gridSize * gridSpacing, 0, i * gridSpacing}
		endPos := Vector3{gridSize * gridSpacing, 0, i * gridSpacing}
		drawWorldLine(startPos, endPos, gridColor, camera, screenWidth, screenHeight, zBuffer)

		// Lines along Z axis
		startPos = Vector3{i * gridSpacing, 0, -gridSize * gridSpacing}
		endPos = Vector3{i * gridSpacing, 0, gridSize * gridSpacing}
		drawWorldLine(startPos, endPos, gridColor, camera, screenWidth, screenHeight, zBuffer)
	}
}

func drawWorldLine(start, end Vector3, color termbox.Attribute, camera Camera, screenWidth, screenHeight int, zBuffer [][]float64) {
	// Project 3D world points to screen space
	startScreen := projectPoint(start, camera, screenWidth, screenHeight)
	endScreen := projectPoint(end, camera, screenWidth, screenHeight)

	// Calculate depths for Z-buffer
	startDepth := worldToCamera(start, camera).Z
	endDepth := worldToCamera(end, camera).Z

	// Draw the line if both points are in front of the camera
	if startDepth > 0 && endDepth > 0 {
		drawLine(startScreen[0], startScreen[1], endScreen[0], endScreen[1],
			color, zBuffer, (startDepth+endDepth)/2)
	}
}

func worldToCamera(worldPoint Vector3, camera Camera) Vector3 {
	// Translate point relative to camera
	relativePt := Vector3{
		X: worldPoint.X - camera.Position.X,
		Y: worldPoint.Y - camera.Position.Y,
		Z: worldPoint.Z - camera.Position.Z,
	}

	// Get camera orientation vectors
	forward := camera.GetDirection()
	right := camera.GetRight()
	up := camera.GetUp()

	// Transform the point to camera space
	return Vector3{
		X: relativePt.X*right.X + relativePt.Y*right.Y + relativePt.Z*right.Z,
		Y: relativePt.X*up.X + relativePt.Y*up.Y + relativePt.Z*up.Z,
		Z: relativePt.X*forward.X + relativePt.Y*forward.Y + relativePt.Z*forward.Z,
	}
}

func projectPoint(worldPoint Vector3, camera Camera, screenWidth, screenHeight int) [2]int {
	// Convert to camera space
	camPoint := worldToCamera(worldPoint, camera)

	// Don't render points behind the camera
	if camPoint.Z <= 0 {
		return [2]int{-1, -1} // Off-screen
	}

	// Simple perspective projection
	scale := 30.0 / camPoint.Z // Larger value = stronger perspective effect
	screenX := int(scale*camPoint.X) + screenWidth/2
	screenY := int(-scale*camPoint.Y) + screenHeight/2

	return [2]int{screenX, screenY}
}

func drawTriangle(triangle Triangle, objectPos Vector3, camera Camera, screenWidth, screenHeight int, zBuffer [][]float64) {
	// Translate triangle vertices to world space
	worldVertices := make([]Vector3, 3)
	for i, v := range triangle.Vertices {
		worldVertices[i] = Vector3{
			X: v.Position.X + objectPos.X,
			Y: v.Position.Y + objectPos.Y,
			Z: v.Position.Z + objectPos.Z,
		}
	}

	// Simple backface culling
	// Calculate normal vector of the triangle
	edge1 := Vector3{
		X: worldVertices[1].X - worldVertices[0].X,
		Y: worldVertices[1].Y - worldVertices[0].Y,
		Z: worldVertices[1].Z - worldVertices[0].Z,
	}
	edge2 := Vector3{
		X: worldVertices[2].X - worldVertices[0].X,
		Y: worldVertices[2].Y - worldVertices[0].Y,
		Z: worldVertices[2].Z - worldVertices[0].Z,
	}

	// Cross product to get normal
	normal := Vector3{
		X: edge1.Y*edge2.Z - edge1.Z*edge2.Y,
		Y: edge1.Z*edge2.X - edge1.X*edge2.Z,
		Z: edge1.X*edge2.Y - edge1.Y*edge2.X,
	}

	// Get the view vector (from triangle to camera)
	viewVector := Vector3{
		X: camera.Position.X - worldVertices[0].X,
		Y: camera.Position.Y - worldVertices[0].Y,
		Z: camera.Position.Z - worldVertices[0].Z,
	}

	// Dot product for culling check
	dotProduct := normal.X*viewVector.X + normal.Y*viewVector.Y + normal.Z*viewVector.Z

	// If dot product is negative, the triangle is facing away from the camera
	if dotProduct < 0 {
		return
	}

	// Project vertices to screen space and calculate depths
	screenVertices := make([][2]int, 3)
	depths := make([]float64, 3)

	for i, v := range worldVertices {
		// Convert world coordinates to camera space
		camSpace := worldToCamera(v, camera)

		// Skip triangles that are behind the camera
		if camSpace.Z <= 0 {
			return
		}

		// Perspective projection
		screenVertices[i] = projectPoint(v, camera, screenWidth, screenHeight)
		depths[i] = camSpace.Z
	}

	// Draw the triangle using a simple rasterization algorithm
	rasterizeTriangle(screenVertices, depths, triangle.Vertices[0].Color, screenWidth, screenHeight, zBuffer)
}

func rasterizeTriangle(vertices [][2]int, depths []float64, color termbox.Attribute, screenWidth, screenHeight int, zBuffer [][]float64) {
	// Find bounding box of the triangle
	minX, minY := screenWidth, screenHeight
	maxX, maxY := 0, 0

	for _, v := range vertices {
		if v[0] < minX {
			minX = v[0]
		}
		if v[0] > maxX {
			maxX = v[0]
		}
		if v[1] < minY {
			minY = v[1]
		}
		if v[1] > maxY {
			maxY = v[1]
		}
	}

	// Clip against screen boundaries
	minX = max(0, minX)
	minY = max(0, minY)
	maxX = min(screenWidth-1, maxX)
	maxY = min(screenHeight-1, maxY)

	// Calculate average depth for z-buffering
	avgDepth := (depths[0] + depths[1] + depths[2]) / 3.0

	// Draw wireframe triangle
	drawLine(vertices[0][0], vertices[0][1], vertices[1][0], vertices[1][1], color, zBuffer, avgDepth)
	drawLine(vertices[1][0], vertices[1][1], vertices[2][0], vertices[2][1], color, zBuffer, avgDepth)
	drawLine(vertices[2][0], vertices[2][1], vertices[0][0], vertices[0][1], color, zBuffer, avgDepth)
}

func drawLine(x0, y0, x1, y1 int, color termbox.Attribute, zBuffer [][]float64, depth float64) {
	dx := abs(x1 - x0)
	dy := abs(y1 - y0)
	sx := 1
	if x0 >= x1 {
		sx = -1
	}
	sy := 1
	if y0 >= y1 {
		sy = -1
	}
	err := dx - dy

	width, height := termbox.Size()

	for {
		if x0 >= 0 && x0 < width && y0 >= 0 && y0 < height {
			// Z-buffer check - only draw if this point is closer than any previously drawn
			if depth < zBuffer[x0][y0] {
				zBuffer[x0][y0] = depth
				termbox.SetCell(x0, y0, getShadeChar(depth), color, termbox.ColorDefault)
			}
		}

		if x0 == x1 && y0 == y1 {
			break
		}

		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x0 += sx
		}
		if e2 < dx {
			err += dx
			y0 += sy
		}
	}
}

// getShadeChar returns a character representing depth (closer points = denser characters)
func getShadeChar(depth float64) rune {
	if depth < 3 {
		return '█'
	} else if depth < 6 {
		return '▓'
	} else if depth < 9 {
		return '▒'
	} else if depth < 12 {
		return '░'
	} else {
		return '·'
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// drawString writes a string to the terminal
func drawString(x, y int, str string, fg, bg termbox.Attribute) {
	for i, c := range str {
		termbox.SetCell(x+i, y, c, fg, bg)
	}
}
