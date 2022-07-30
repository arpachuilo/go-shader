#version 410

uniform vec2 u_resolution;
uniform vec2 u_mouse;
uniform float u_time;
uniform int u_frame;

in vec2 fragTexCoord;

out vec4 outputColor;
#define PI 3.14159265359

// Plot a line on Y using a value between 0.0-1.0
float plot(vec2 st, float pct){
  return  smoothstep( pct-0.2, pct, st.y) -
          smoothstep( pct, pct+0.2, st.y);
}

void main() {
  vec2 st = gl_FragCoord.xy/u_resolution;

  // float y = st.x;
  // float y = pow(st.x, PI);
  // float y = step(0.5,st.x);
  // float y = smoothstep(0.2,0.8,st.x);
  // float y = smoothstep(0.1,0.5,st.x) - smoothstep(0.5,0.9,st.x);
  float x = 2*u_time + st.x * PI * 2;
  // float y = sin(u_time * st.x + st.x * PI);

  // y = abs(y);
  // y = fract(y);
  // y = ceil(y) + floor(y); // scanlines with high freq.

  float y = sin(x) / 4 + 0.5;
  // y = mod(x,0.5); // return x modulo of 0.5
  // y = fract(x); // return only the fraction part of a number
  // y = ceil(x);  // nearest integer that is greater than or equal to x
  // y = floor(x); // nearest integer less than or equal to x
  // y = sign(x);  // extract the sign of x
  // y = abs(x);   // return the absolute value of x
  // y = clamp(x,0.0,1.0); // constrain x to lie between 0.0 and 1.0
  // y = min(0.0,x);   // return the lesser of x and 0.0
  // y = max(0.0,x);   // return the greater of x and 0.0
  vec3 color = vec3(y);

  // Plot a line
  float pct = plot(st, y);
  color = 
    (1.0 - pct) * color
    + pct * vec3(
      sin(u_time),
      cos(u_time),
      tan(u_time)
    );

	outputColor = vec4(color,1.0);
}
