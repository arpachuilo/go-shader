#version 410

uniform vec2 u_resolution;
uniform vec2 u_mouse;
uniform float u_time;
uniform int u_frame;
uniform float u_farclip;

uniform sampler2D p_buffer;
uniform int p_use;

in vec4 fragTexCoord;
in vec4 ex_position;

out vec4 outputColor;

vec3 turbo(float x) {
  float r = 0.1357 + x * ( 4.5974 - x * ( 42.3277 - x * ( 130.5887 - x * ( 150.5666 - x * 58.1375 ))));
  float g = 0.0914 + x * ( 2.1856 + x * ( 4.8052 - x * ( 14.0195 - x * ( 4.2109 + x * 2.7747 ))));
  float b = 0.1067 + x * ( 12.5925 - x * ( 60.1097 - x * ( 109.0745 - x * ( 88.5066 - x * 26.8183 ))));
  return vec3(r,g,b);
}

vec4 get(vec2 coord) {
  vec3 t = texture(p_buffer, vec2(gl_FragCoord.xy + coord) / u_resolution, 0).xyz;
  return vec4(t.zyx, ex_position.w);
}

void main() {
  outputColor = vec4(turbo(1.0 - ex_position.w/u_farclip), 1.0);
  // outputColor = vec4(1.0);
}
