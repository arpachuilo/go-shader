#version 410
uniform sampler2D state;

uniform float cursorSize;
uniform float time;
uniform vec2 scale;
uniform vec2 mouse;

uniform vec2 particles[100]; // particles
uniform float particleSizes[100];
uniform int len;

in vec2 fragTexCoord;

out vec4 outputColor;

vec4 uv() {
  return texture(state, vec2(gl_FragCoord.xy) / scale, 0);
}

// line renderer
void main() {

  vec2 c = gl_FragCoord.xy;
  for (int i = 0; i < len; i++) {
    if (length(c-particles[i]) < particleSizes[i]) {
      outputColor = vec4(1.0);
      return;
    }
  }
  
  // decay previous areas
  outputColor = uv() - 0.01;
}
