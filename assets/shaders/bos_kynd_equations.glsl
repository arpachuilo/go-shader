// Exercise from book of shaders
// kynd's table of equations https://thebookofshaders.com/05/kynd.png
#version 410

uniform vec2 u_resolution;
uniform vec2 u_mouse;
uniform float u_time;
uniform int u_frame;

in vec2 fragTexCoord;

out vec4 outputColor;
#define PI 3.14159265359

// mappers
float map(float value, float inMin, float inMax, float outMin, float outMax) {
  return outMin + (outMax - outMin) * (value - inMin) / (inMax - inMin);
}

vec2 map(vec2 value, vec2 inMin, vec2 inMax, vec2 outMin, vec2 outMax) {
  return outMin + (outMax - outMin) * (value - inMin) / (inMax - inMin);
}

vec3 map(vec3 value, vec3 inMin, vec3 inMax, vec3 outMin, vec3 outMax) {
  return outMin + (outMax - outMin) * (value - inMin) / (inMax - inMin);
}

vec4 map(vec4 value, vec4 inMin, vec4 inMax, vec4 outMin, vec4 outMax) {
  return outMin + (outMax - outMin) * (value - inMin) / (inMax - inMin);
}

// Plot a line on Y using a value between 0.0-1.0
float plot(vec2 st, float pct) {
  return 
    smoothstep(
      0.7, 1.0,
      smoothstep(pct-0.1, pct, st.y) 
      - smoothstep(pct, pct+0.01, st.y)
    );
}

vec3 color_fg(float pct, vec3 current, vec3 color) {
  // return (1.0 - step(pct, 0.9)) * color + step(pct, 0.1) * current;
  return (1.0 - pct) * current + pct * color;
}

void main() {
  vec2 st = gl_FragCoord.xy/u_resolution;

  float speed = u_time * 1.0;

  // movement along x
  // float x = st.x; // none
  // float x = sin(2*speed + st.x * PI) / 2 + 0.5; // focus center
  float x = sin(speed + st.x * PI); // full


  // bg color
  vec3 color = vec3(0.0);

  // movement of exp
  float e = map(sin(2*speed), -1.0, 1.0, 0.5, 3.5);
  // f1(x, e)
  color = color_fg(
    plot(st, 1.0 - pow(abs(x), e)),
    color, vec3(0.0, 1.0, 0.0)
  );

  // f2(x, e)
  color = color_fg(
    plot(st, pow(cos(PI * x / 2.0), e)),
    color, vec3(1.0, 0.0, 0.0)
  );

  // f3(x, e)
  color = color_fg(
    plot(st, 1.0 - pow(abs(sin(PI * x / 2.0)), e)),
    color, vec3(0.0, 0.0, 1.0)
  );

  // f4(x, e)
  color = color_fg(
    plot(st, pow(min(cos(PI * x / 2.0), 1.0 - abs(x)), e)),
    color, vec3(0.0, 1.0, 1.0)
  );

  // f5(x, e)
  color = color_fg(
    plot(st, pow(max(0.0, abs(x) * 2.0 - 1.0), e)),
    color, vec3(1.0, 1.0, 0.0)
  );

	outputColor = vec4(color,1.0);
}
