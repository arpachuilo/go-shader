#version 410

uniform vec2 u_resolution;
uniform vec2 u_mouse;
uniform float u_time;
uniform int u_frame;

layout(location = 0) in vec2 pos;
layout(location = 1) in vec2 tex;

out vec2 fragTexCoord;

void main() {
  gl_Position = vec4(pos, 0, 1);
  fragTexCoord = tex;
}

