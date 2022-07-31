// sample state using uv and apply turbo coloring
#version 410
uniform int index;
uniform sampler2D state;
uniform vec2 scale;
uniform float alpha;


vec3 turbo(float x) {
  float r = 0.1357 + x * ( 4.5974 - x * ( 42.3277 - x * ( 130.5887 - x * ( 150.5666 - x * 58.1375 ))));
  float g = 0.0914 + x * ( 2.1856 + x * ( 4.8052 - x * ( 14.0195 - x * ( 4.2109 + x * 2.7747 ))));
  float b = 0.1067 + x * ( 12.5925 - x * ( 60.1097 - x * ( 109.0745 - x * ( 88.5066 - x * 26.8183 ))));
  return vec3(r,g,b);
}

in vec2 fragTexCoord;

out vec4 outputColor;

void main() {
  vec4 tex = texture(state, gl_FragCoord.xy  / scale, 0);
  vec4 color = vec4(turbo(tex[index]), 1.0);
  outputColor = vec4(color.rgb, 1.0 - color.a * alpha);
}
