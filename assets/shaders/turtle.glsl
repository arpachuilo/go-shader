// turtle drawer
#version 410
uniform sampler2D state;

uniform float cursorSize;
uniform float time;
uniform int d;
uniform float w;
uniform vec2 a;
uniform vec2 b;
uniform vec2 scale;
uniform vec2 mouse;

in vec2 fragTexCoord;

out vec4 outputColor;

vec4 uv() {
  return texture(state, vec2(gl_FragCoord.xy) / scale, 0);
}

// line renderer
void main() {
  if (d == 0) {
    outputColor = uv();
    return;
  }

  // within segment
  vec2 c = gl_FragCoord.xy;
  vec2 u = normalize(b - a);
  vec2 v = vec2(-u.y, u.x);
  float du = dot(c - a, u);
  if (0 <= du && du <= length(b - a)) {
    float dv = dot(c - a, v);
    if (abs(dv) < w / 2) {
      // float g = abs(du) / (w / 2);
      // if (g <= 0.5) {
      //   g = 1.0;
      // } 

      // outputColor = vec4(g);
      outputColor = vec4(1.0);
      return;
    }
  }

  outputColor = uv();
}
