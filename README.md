# OpenGL Things

Learning OpenGL and GLFW using Go.

## General Keybinds

* `KeyF1` - Unlock framerate
* `KeyF2` - Screenshot
* `KeyF3` - Record to .avi

## Game of Life Shader

Game of life shader.
![game_of_life](https://user-images.githubusercontent.com/8808952/188760352-218303b0-d106-4bd3-93d9-f3b98edf29bc.png)
![game_of_life_2](https://user-images.githubusercontent.com/8808952/188760364-b376103b-1fad-4c8b-a336-5be01db35883.png)
![cyclic_life_2](https://user-images.githubusercontent.com/8808952/188760385-6ab3ad32-f0a1-404c-a02e-e6da5fa9dd66.png)
![cyclic_life](https://user-images.githubusercontent.com/8808952/188760393-4155045d-1ee5-4135-809a-eda472a7d9a4.png)

`make run PROGRAM=game_of_life`

* `Key1` - Switch to standard life mode (default)
* `Key2` - Switch to cyclife life
* `KeyJ` / `KeyK` - Cycle coloring used
* `KeyH` / `KeyL` - Cycle channel
* `KeySpace` - Pause simulation

## Smooth Life Shader

Smooth life shader.

`make run PROGRAM=smooth_life`

* `KeyJ` / `KeyK` - Cycle coloring used
* `KeyH` / `KeyL` - Cycle channel
* `KeySpace` - Pause simulation

## Julia Shader

Julia fractal shader.

`make run PROGRAM=julia`

* `KeyJ` / `KeyK` - Cycle coloring used
* `KeyEqual` - Increase iterations.
* `KeyMinus` - Decrease iterations.

## Mandelbrot Shader

Mandelbrot fractal shader.

`make run PROGRAM=mandelbrot`

* `KeyJ` / `KeyK` - Cycle coloring used
* `KeyEqual` - Increase iterations.
* `KeyMinus` - Decrease iterations.

## Pong Shader

Some bouncing balls with a trail

`make run PROGRAM=pong`

* `KeyJ` / `KeyK` - Cycle coloring used

## Turtle Graphics Shader

Turtle graphics shader I started work on

`make run PROGRAM=turtle`

* `KeyJ` / `KeyK` - Cycle coloring used
