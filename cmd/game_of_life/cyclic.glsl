// cyclic life frag shader
#version 410
uniform float stages;
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

vec4 nextStage(float e, vec4 current) {
  vec4 next = current + vec4(1.0 / stages);

  return vec4(
    abs(next.r - 1.0) < e ? 0.0 : next.r,
    abs(next.g - 1.0) < e ? 0.0 : next.g,
    abs(next.b - 1.0) < e ? 0.0 : next.b,
    abs(next.a - 1.0) < e ? 0.0 : next.a
  );
}

ivec4 successor(float e, vec4 next, vec4 neighbors) {
    return ivec4(
      abs(neighbors.r - next.r) < e ? 1 : 0,
      abs(neighbors.g - next.g) < e ? 1 : 0,
      abs(neighbors.b - next.b) < e ? 1 : 0,
      abs(neighbors.a - next.a) < e ? 1 : 0
    );
}

vec4 op(vec4 current, vec4 next, ivec4 neighbors) {
  return vec4(
    neighbors.r > 0 ? next.r : current.r, 
    neighbors.g > 0 ? next.g : current.g, 
    neighbors.b > 0 ? next.b : current.b, 
    neighbors.a > 0 ? next.a : current.a
  );
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

  float e = 1.0 / stages / 2.0;
  vec4 current = get(vec2(0, 0));
  vec4 next = nextStage(e, current);
  ivec4 neighbors = successor(e, next, get(vec2(-1,  0))) +
                    successor(e, next, get(vec2( 0, -1))) +
                    successor(e, next, get(vec2( 0,  1))) +
                    successor(e, next, get(vec2( 1,  0)));

  outputColor = op(current, next, neighbors);
}
