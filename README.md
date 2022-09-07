# OpenGL Things

Learning OpenGL and GLFW using Go.

## General Keybinds

* `KeyF1` - Unlock framerate
* `KeyF2` - Screenshot
* `KeyF3` - Record to .avi

## Game of Life Shader

Game of life shader.
`make run PROGRAM=game_of_life`

* `Key1` - Switch to standard life mode (default)
* `Key2` - Switch to cyclife life
* `KeyJ` / `KeyK` - Cycle coloring used
* `KeyH` / `KeyL` - Cycle channel
* `KeySpace` - Pause simulation

Images from standard game of life
![game_of_life](https://user-images.githubusercontent.com/8808952/188760352-218303b0-d106-4bd3-93d9-f3b98edf29bc.png)
![game_of_life_2](https://user-images.githubusercontent.com/8808952/188760364-b376103b-1fad-4c8b-a336-5be01db35883.png)

Images form cyclic game of life
![cyclic_life_2](https://user-images.githubusercontent.com/8808952/188760385-6ab3ad32-f0a1-404c-a02e-e6da5fa9dd66.png)
![cyclic_life](https://user-images.githubusercontent.com/8808952/188760393-4155045d-1ee5-4135-809a-eda472a7d9a4.png)

## Smooth Life Shader

Smooth life shader.

`make run PROGRAM=smooth_life`

* `KeyJ` / `KeyK` - Cycle coloring used
* `KeyH` / `KeyL` - Cycle channel
* `KeySpace` - Pause simulation

Image from smooth life shader
![smooth_life](https://user-images.githubusercontent.com/8808952/188760614-108d69e4-4994-457d-a92e-199ec9493570.png)

## Julia Shader

Julia fractal shader.

`make run PROGRAM=julia`

* `KeyJ` / `KeyK` - Cycle coloring used
* `KeyEqual` - Increase iterations.
* `KeyMinus` - Decrease iterations.

Image from julia shader
![julia](https://user-images.githubusercontent.com/8808952/188760658-ab8cc850-abbd-404e-a4e1-8dd3a8bf5075.png)

## Mandelbrot Shader

Mandelbrot fractal shader.

`make run PROGRAM=mandelbrot`

* `KeyJ` / `KeyK` - Cycle coloring used
* `KeyEqual` - Increase iterations.
* `KeyMinus` - Decrease iterations.

Image from mandelbrot shader
![mandelbrot_2](https://user-images.githubusercontent.com/8808952/188760709-8699500f-edb2-49c2-8380-f2ba122ffac0.png)

## Pong Shader

Some bouncing balls with a trail

`make run PROGRAM=pong`

* `KeyJ` / `KeyK` - Cycle coloring used

Image from pong shader
![pong](https://user-images.githubusercontent.com/8808952/188760732-07fa0ed1-b76e-45e6-b51c-2c07a1b7fa6a.png)

## Turtle Graphics Shader

Turtle graphics shader I started work on

`make run PROGRAM=turtle`

* `KeyJ` / `KeyK` - Cycle coloring used

Image from turtle shader
![turtle](https://user-images.githubusercontent.com/8808952/188760891-be8e6f3d-0463-44ef-9bc2-a9f97b12b691.png)

## Shader Watch

Shaders reload when edited. First work on 3D.

`make run PROGRAM=shader_watch`

![cubes](https://user-images.githubusercontent.com/8808952/188760991-30d50a70-4ef6-4978-9b8b-fb3ca83d2b33.png)

