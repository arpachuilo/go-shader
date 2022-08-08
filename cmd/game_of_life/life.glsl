#version 410
uniform int s[9];
uniform int b[9];
uniform sampler2D state;

uniform float cursorSize;
uniform float u_time;
uniform vec2 u_resolution;
uniform vec2 u_mouse;

in vec2 fragTexCoord;

out vec4 outputColor;

vec4 get(vec2 coord) {
  return texture(state, vec2(gl_FragCoord.xy + coord) / u_resolution, 0);
}

float random (vec2 st) {
  return fract(sin(dot(st.xy, vec2(12.9898,78.233)))*43758.5453123);
}

ivec4 alive(vec4 cell) {
  return ivec4(step(0.5, cell));
}

float op(float c, int n) {
  if (
    n == b[0] || 
    n == b[1] || 
    n == b[2] || 
    n == b[3] || 
    n == b[4] || 
    n == b[5] || 
    n == b[6] || 
    n == b[7] || 
    n == b[8]
  ) {
    return 1.0;
  } else if (
    n == s[0] || 
    n == s[1] || 
    n == s[2] || 
    n == s[3] || 
    n == s[4] || 
    n == s[5] || 
    n == s[6] || 
    n == s[7] || 
    n == s[8]
  ) {
    return c;
  }

  return 0.0;
}

void main() {
  vec2 pos = gl_FragCoord.xy;
  if (u_mouse.x < (0.01 * u_resolution.x) && u_time > 1) {
    outputColor = vec4(0.0);
    return;
  } else if (u_mouse.x > (0.99 * u_resolution.x) || length(pos-u_mouse) < (cursorSize * u_resolution.x)) {
    outputColor = vec4(
      random(pos / u_resolution),
      random(u_mouse / u_resolution),
      random(vec2(sin(u_time), sin(u_time)) / u_resolution),
      random(vec2(cos(u_time), cos(u_time)) / u_resolution)
    );
    return;
  }

  // sum each channel alive
  float oo = 1.0; // makes cool effect this way
  ivec4 sum = alive(get(vec2(-oo, -oo))) +
              alive(get(vec2(-oo,  0))) +
              alive(get(vec2(-oo,  oo))) +
              alive(get(vec2( 0, -oo))) +
              alive(get(vec2( 0,  oo))) +
              alive(get(vec2( oo, -oo))) +
              alive(get(vec2( oo,  0))) +
              alive(get(vec2( oo,  oo)));

  vec4 current = get(vec2(0, 0));
  outputColor = vec4(
      op(current.r, sum.r),
      op(current.g, sum.g),
      op(current.b, sum.b),
      op(current.a, sum.a)
  );
}
