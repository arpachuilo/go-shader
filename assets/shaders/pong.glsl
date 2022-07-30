#version 410
uniform sampler2D iChannel1;

uniform float iTime;
uniform vec2 iResolution;
uniform vec2 iMouse;

uniform vec2 pPos[100]; // particles
uniform float pSize[100];
uniform int len;

in vec2 fragTexCoord;

out vec4 outputColor;

vec4 uv() {
  return texture(iChannel1, vec2(gl_FragCoord.xy) / iResolution, 0);
}

// line renderer
void main() {

  vec2 c = gl_FragCoord.xy;
  for (int i = 0; i < len; i++) {
    float l = length(c - pPos[1]);
    if (length(c-pPos[i]) < pSize[i]) {
      outputColor = vec4(1.0);
      return;
    }
  }
  
  // decay previous areas
  vec4 oc = uv();
  outputColor = vec4(oc.rgb - 0.01, step(0.1, oc.r));
}
