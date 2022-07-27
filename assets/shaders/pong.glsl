#version 410
uniform sampler2D state;

uniform float cursorSize;
uniform float time;
uniform float size;
uniform vec2 scale;
uniform vec2 mouse;

uniform vec2 b; // particle location

in vec2 fragTexCoord;

out vec4 outputColor;

vec4 uv() {
  return texture(state, vec2(gl_FragCoord.xy) / scale, 0);
}

// line renderer
void main() {

  vec2 c = gl_FragCoord.xy;
  if (length(c-b) < size) {
    outputColor = vec4(1.0);
    return;
  }

  // decay previous areas
  outputColor = uv() - 0.01;
}
