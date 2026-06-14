# TermiGoCraft

A first-person, Minecraft-style **voxel world rendered in the terminal**, in Go.

The world is drawn by a per-pixel **voxel raycaster** (Amanatides & Woo DDA); entities,
the held item, and the block-selection highlight are drawn by a **triangle rasterizer**.
Both write into one shared **color + perpendicular-depth framebuffer**, so occlusion
between terrain and entities is automatically correct. Output uses
[`tcell/v2`](https://github.com/gdamore/tcell) with **truecolor half-block (`▀`) pixels**
(two pixels per terminal cell). Tested in **kitty**.

## Build & run

```bash
go build -o termigocraft .   # or: make build
./termigocraft               # or: make run / go run .
```

Requires Go 1.24+ and a truecolor terminal (kitty recommended; any terminal that
sets `COLORTERM=truecolor` works, others fall back to the nearest palette).

## Controls

| Keys | Action |
|------|--------|
| `W` `A` `S` `D` | move (relative to facing) |
| `Space` / `C` | move up / down |
| Arrow keys | look around |
| `F` / `Enter` | break the targeted block |
| `R` | place the selected block |
| `Q` / `E` or `1`–`9` | select block type |
| `H` / `?` | toggle help |
| `Esc` / `Ctrl+C` | quit |

## Design

- **`internal/geom`** — `Vec3`, `RGB`, `Triangle` value types.
- **`internal/world`** — flat `[]Block` voxel grid, terrain generation, DDA `Cast`, break/place edits.
- **`internal/camera`** — first-person camera (position + yaw/pitch), basis vectors, per-pixel rays.
- **`internal/render`** — shared framebuffer, the world raycaster, the triangle rasterizer, the HUD.
- **`internal/entity`** — non-block meshes (and the held item).
- **`internal/input`** — key → action mapping.
- **`internal/game`** — state + the fixed-timestep loop (input on a goroutine, ~30 FPS).

### Performance / memory

The steady-state frame performs **zero heap allocations**: all buffers are
preallocated and reused, the world is a flat pointer-free `[]uint8`, and `Present`
skips unchanged cells (so an idle view touches the terminal not at all). Guarded by
benchmarks:

```bash
make bench    # BenchmarkFrame and BenchmarkClearPresent -> 0 B/op, 0 allocs/op
```

## Tests

```bash
make test     # geom, world (DDA + edits), camera, input, render (occlusion), game pipeline
make vet
```

Spec-driven design docs live in [`specs/001-first-person-voxel-world/`](specs/001-first-person-voxel-world/).
