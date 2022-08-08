#version 410

#define PI 3.14159265359
#define TAU 6.28318530718
#define cbrtf(x)  (sign(x)*pow(abs(x),1./3.))

uniform vec2 u_resolution;
uniform vec2 u_mouse;
uniform float u_time;
uniform int u_frame;
uniform float u_farclip;

in vec4 fragTexCoord;
in vec4 ex_position;

layout(location = 0) out vec4 outputColor;

vec3 rgb2hsb( in vec3 c ){
    vec4 K = vec4(0.0, -1.0 / 3.0, 2.0 / 3.0, -1.0);
    vec4 p = mix(vec4(c.bg, K.wz),
                 vec4(c.gb, K.xy),
                 step(c.b, c.g));
    vec4 q = mix(vec4(p.xyw, c.r),
                 vec4(c.r, p.yzx),
                 step(p.x, c.r));
    float d = q.x - min(q.w, q.y);
    float e = 1.0e-10;
    return vec3(abs(q.z + (q.w - q.y) / (6.0 * d + e)),
                d / (q.x + e),
                q.x);
}

//  Function from Iñigo Quiles
//  https://www.shadertoy.com/view/MsS3Wc
vec3 hsb2rgb( in vec3 c ){
    vec3 rgb = clamp(abs(mod(c.x*6.0+vec3(0.0,4.0,2.0),
                             6.0)-3.0)-1.0,
                     0.0,
                     1.0 );
    rgb = rgb*rgb*(3.0-2.0*rgb);
    return c.z * mix(vec3(1.0), rgb, c.y);
}
float gain(float x, float k) {
  float a = 0.5*pow(2.0*((x<0.5)?x:1.0-x), k);
  return (x<0.5)?a:1.0-a;
}
float parabola(float x, float k) {
  return pow(4.0*x*(1.0-x), k);
}

float cubicPulse(float c, float w, float x) {
  x = abs(x - c);
  if(x > w) return 0.0;
  x /= w;
  return 1.0 - x*x*(3.0-2.0*x);
}

float plot (vec2 st, float pct) {
  return  smoothstep( pct-0.01, pct, st.y) -
          smoothstep( pct, pct+0.01, st.y);
}

float plot_alt(vec2 st, float pct) {
  return cubicPulse(pct, 0.005, st.y);
}

vec3 turbo(float x) {
  float r = 0.1357 + x * ( 4.5974 - x * ( 42.3277 - x * ( 130.5887 - x * ( 150.5666 - x * 58.1375 ))));
  float g = 0.0914 + x * ( 2.1856 + x * ( 4.8052 - x * ( 14.0195 - x * ( 4.2109 + x * 2.7747 ))));
  float b = 0.1067 + x * ( 12.5925 - x * ( 60.1097 - x * ( 109.0745 - x * ( 88.5066 - x * 26.8183 ))));
  return vec3(r,g,b);
}

vec3 rgb2oklab(vec3 c) {
  float l = 0.4122214708f * c.r + 0.5363325363f * c.g + 0.0514459929f * c.b;
  float m = 0.2119034982f * c.r + 0.6806995451f * c.g + 0.1073969566f * c.b;
  float s = 0.0883024619f * c.r + 0.2817188376f * c.g + 0.6299787005f * c.b;

  float l_ = cbrtf(l);
  float m_ = cbrtf(m);
  float s_ = cbrtf(s);

  return vec3(
    0.2104542553f*l_ + 0.7936177850f*m_ - 0.0040720468f*s_,
    1.9779984951f*l_ - 2.4285922050f*m_ + 0.4505937099f*s_,
    0.0259040371f*l_ + 0.7827717662f*m_ - 0.8086757660f*s_
  );
}

vec3 oklab2rgb(vec3 c) {
  float l_ = c.x + 0.3963377774f * c.y + 0.2158037573f * c.z;
  float m_ = c.x - 0.1055613458f * c.y - 0.0638541728f * c.z;
  float s_ = c.x - 0.0894841775f * c.y - 1.2914855480f * c.z;

  float l = l_*l_*l_;
  float m = m_*m_*m_;
  float s = s_*s_*s_;

  float r = +4.0767416621f * l - 3.3077115913f * m + 0.2309699292f * s;
  float g = -1.2684380046f * l + 2.6097574011f * m - 0.3413193965f * s;
  float b = -0.0041960863f * l - 0.7034186147f * m + 1.7076147010f * s;

  return vec3(r, g, b);
}

// void main() {
//     outputColor = vec4(turbo(1.0 - ex_position.w/u_farclip*2), 1.0);
// }

// https://thebookofshaders.com/06/
// # Playing with gradients
// void main() {
//   vec2 st = gl_FragCoord.xy/u_resolution.xy;
//
//   float t = 1*u_time;
//   float x = st.x * PI * 1;
//   vec3 pct = vec3(st.x);
//   // vec3 pct = vec3(st.x);
//
//   pct.r = smoothstep(0.0, 1.0, sin(t + x));
//   pct.g = sin(st.x*PI);
//   // pct.b = gain(sin(t + -x), sin(t + -x));
//   pct.b = gain(2.0, sin(t + -x));
//
//   vec3 color = vec3(0.0);
//   vec3 colorA = vec3(0.149,0.141,0.792);
//   vec3 colorB = vec3(1.000,0.333,0.224);
//
//   color = mix(colorA, colorB, pct);
//
//   // Plot transition lines for each channel
//   color = mix(color, vec3(1.0,0.0,0.0), plot_alt(st, pct.r));
//   color = mix(color, vec3(0.0,1.0,0.0), plot_alt(st, pct.g));
//   color = mix(color, vec3(0.0,0.0,1.0), plot_alt(st, pct.b));
//
//   outputColor = vec4(color,1.0);
// }

// # HSB
// void main(){
//   vec2 st = gl_FragCoord.xy/u_resolution;
//   vec3 color = vec3(0.0);
//   float t = 1*u_time;
//   float x = st.x * PI * 1;
//   // We map x (0.0 - 1.0) to the hue (0.0 - 1.0)
//   // And the y (0.0 - 1.0) to the brightness
//   color = hsb2rgb(vec3(st.x,1.0,st.y));
//
//   outputColor = vec4(color,1.0);
// }

// # HSB in polar coordinates
void main() {
  // vec2 st = gl_FragCoord.xy/u_resolution;
  vec2 st = fragTexCoord.xy;
  vec3 color = vec3(0.0);

  // Use polar coordinates instead of cartesian
  // chroma
  // vec2 C = vec2(0.5)-st;
  // float angle = atan(C.y, C.x); 
  // float radius = length(C)*2.0;
  // float h = angle/TAU + 0.5; // without rotate animation
  // float h = angle/TAU + fract(u_time/4); // with rotate animation
  // color = hsb2rgb(vec3(h,radius,1.0));

  // map from oklab
  // float h = angle/TAU; // with rotate animation
  vec2 P = vec2(0.5)-st; // origin
  float radius = length(P)*2.0;
  float C = distance(vec2(0.5), st); // Chrome
  float h = atan(P.y, P.x); // hue
  // h = sin(h + u_time)*2.0;
  // h = h + fract(u_time/4);
  // float a = C*(cos(h) + sin(u_time/2)); // a green/red
  float a = C*cos(h); // a green/red
  // float b = C*(sin(h) + cos(u_time/2)); // b blue/yellow
  float b = C*sin(h); // b blue/yellow
  color = oklab2rgb(vec3(step(radius, 0.8), a, b));


  // float t = sin(u_time/1);
  // color.r = gain(t, radius);
  // color.r = smoothstep(0.3, 0.7, sin(radius));
  // color.g = smoothstep(0.0, 0.8, cos(radius));
  // color.b = smoothstep(0.0, 0.8, sin(radius));

  outputColor = vec4(color,1.0);
}
