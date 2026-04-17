# icosa-gravity: N-Body Cubic Cluster

A high-performance 3D N-Body gravitational simulation written in Go using OpenGL 4.1. This project simulates thousands of cubic particles influenced by mutual gravitational attraction, optimized with the legendary Fast Inverse Square Root (Quake III) algorithm and multithreaded CPU physics.

## 🚀 Demo

![N-Body Cluster Simulation](./Raster.gif)

*The simulation starts as a spherical shell and collapses into complex filaments and clusters based on mutual gravitational pull.*

## 🛠 Features

1. **N-Body Gravity Physics:** Every particle exerts a gravitational force on every other particle ($O(N^2)$ complexity).
2. **Fast Inverse Square Root:** Uses the `0x5f3759df` bit-level hack for rapid $1/r^3$ force calculations.
3. **Instanced Rendering:** Efficiently draws thousands of 3D cubes by sending geometry to the GPU once and instancing positions.
4. **Multithreaded Physics:** Distributes physics calculations across multiple CPU workers using Go's concurrency primitives (`sync.WaitGroup`).
5. **Post-Processing:** Custom Framebuffer Object (FBO) implementation with a brightness-boosting composite pass.
6. **3D Camera:** Fully interactive Fly-cam with mouse look and keyboard navigation.

## ⌨️ Controls

| Key | Action |
| :--- | :--- |
| **W / S** | Move Forward / Backward |
| **A / D** | Strafe Left / Right |
| **Mouse Drag** | Look around (Left Click + Move) |
| **Scroll** | Increase/Decrease movement speed |
| **P / O** | Increase/Decrease Cube Size |
| **R** | Reset Simulation (Re-spawn sphere) |

## 📦 Installation & Running

### Prerequisites
* **Go 1.20+**
* **C Compiler** (GCC or Clang for GLFW/OpenGL bindings)
* **OpenGL 4.1** compatible hardware

### Setup
1. Clone the repository:
   ```bash
   git clone [https://github.com/yourusername/icosa-gravity.git](https://github.com/yourusername/icosa-gravity.git)
   cd icosa-gravity



### Install dependencies:

```bash
go mod init icosa-gravity
go get [github.com/go-gl/gl/v4.1-core/gl](https://github.com/go-gl/gl/v4.1-core/gl)
go get [github.com/go-gl/glfw/v3.3/glfw](https://github.com/go-gl/glfw/v3.3/glfw)
go get [github.com/go-gl/mathgl/mgl32](https://github.com/go-gl/mathgl/mgl32)
