#version 410

uniform vec2 u_resolution;
uniform vec2 u_mouse;
uniform float u_time;
uniform int u_frame;

in vec2 vert;
in vec2 vertTexCoord;

out vec2 fragTexCoord;

void main() {
  gl_Position = vec4(vert, 0, 1);
  fragTexCoord = vertTexCoord;
}
