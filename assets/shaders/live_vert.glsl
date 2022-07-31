#version 410

uniform vec2 u_resolution;
uniform vec2 u_mouse;
uniform float u_time;
uniform int u_frame;

in vec2 vert;
in vec2 vertTexCoord;

out vec2 fragTexCoord;

float random (vec2 st) {
  return fract(sin(dot(st.xy,
                       vec2(12.9898,78.233)))*
      43758.5453123);
}

void main() {
  vec2 st = vert/u_resolution.xy;
  float rnd = random(st) * 10;

  vec2 tt = vertTexCoord;
  tt.x += sin(u_time) * 1.0;
  // tt.x = fract(tt.x);
  // tt.y += cos(u_time + rnd) * 10;
  // gl_Position = vec4(tt, 0, 1);
  // fragTexCoord = tt;

  gl_Position = vec4(vert, 0, 1);
  fragTexCoord = vertTexCoord;
  // fragTexCoord = vert;
}
