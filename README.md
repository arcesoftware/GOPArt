# icosa-gravity: N-Body Cubic Cluster

A high-performance 3D N-body gravitational simulation written in Go using OpenGL 4.1.  
This project simulates thousands of cubic particles interacting via gravity, forming emergent structures like clusters and filaments.

---

## 🚀 Demo

![N-Body Cluster Simulation](./Raster.gif)

The simulation starts as a spherical distribution and evolves dynamically under mutual gravitational attraction.

---

## 🧠 Features

- **N-Body Gravity (O(N²))**  
  Every particle interacts with every other particle using Newtonian gravity.

- **Fast Inverse Square Root**  
  Uses the classic Quake III `0x5f3759df` trick for fast inverse distance calculations.

- **Instanced Rendering (GPU Efficient)**  
  Thousands of cubes rendered using a single mesh with instancing.

- **Multithreaded Physics Engine**  
  Physics updates are parallelized using Go routines and `sync.WaitGroup`.

- **Post-Processing Pipeline**  
  Framebuffer Object (FBO) with a composite pass for brightness enhancement.

- **3D Fly Camera**  
  Smooth WASD movement + mouse look.

---

## ⌨️ Controls

| Key | Action |
|-----|--------|
| W / S | Move Forward / Backward |
| A / D | Strafe Left / Right |
| Mouse (Hold Left Click) | Look Around |
| Scroll | Adjust Movement Speed |
| P / O | Increase / Decrease Cube Size |
| R | Reset Simulation |

---

## ⚙️ Installation & Running

### Prerequisites

- Go 1.20+
- C Compiler (GCC or Clang)
- OpenGL 4.1 compatible GPU

---

### Setup

```bash
git clone https://github.com/arcesoftware/GOPArt.git
cd GOPArt
```

### Install Dependencies

```bash
go mod init GOPArt
go get github.com/go-gl/gl/v4.1-core/gl
go get github.com/go-gl/glfw/v3.3/glfw
go get github.com/go-gl/mathgl/mgl32
```

---

### Run

```bash
go run main.go
```

---

## 🧪 Performance Notes

- Complexity: **O(N²)**
- Recommended:
  - ~2,000 particles → smooth
  - ~5,000 particles → moderate load
  - 10,000+ → CPU heavy

---

## ⚠️ Known Limitations

- No spatial partitioning (Barnes-Hut not implemented)
- CPU-bound physics
- No collision handling
- Possible instability at very small distances

---

## 🔧 Stability Tips

To avoid simulation explosions, use softening:

```go
const softening = 25.0
distSqr := dir.Dot(dir) + softening
```

---

## 🔥 Roadmap

- [ ] Barnes-Hut Octree (O(N log N))
- [ ] GPU compute shaders
- [ ] CUDA acceleration
- [ ] Volumetric nebula integration
- [ ] Motion trails / bloom
- [ ] Galaxy formation simulation

---

## 🧬 Inspiration

- Astrophysics simulations  
- Real-time rendering engines  
- Quake III Arena optimization techniques  

---

## 📜 License

MIT License

---

## 👨‍💻 Author

Juan Arce  
Costa Rica 🇨🇷

---

## ⭐ Support

If you like this project:

- ⭐ Star the repo  
- 🍴 Fork it  
- 🧪 Experiment with it  
