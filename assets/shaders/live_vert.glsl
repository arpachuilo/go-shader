#version 410

uniform vec2 u_resolution;
uniform vec2 u_mouse;
uniform float u_time;
uniform int u_frame;

uniform mat4 ModelMatrix;
uniform mat4 ViewMatrix;
uniform mat4 ProjectionMatrix;

in vec4 vert;
in vec4 vertTexCoord;

out vec4 ex_position;
out vec4 fragTexCoord;

void main() {
  gl_Position = (ProjectionMatrix * ViewMatrix * ModelMatrix) * vert;
  ex_position = gl_Position;
  fragTexCoord = vertTexCoord;
}
