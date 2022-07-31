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

// useful little functions https://iquilezles.org/articles/functions/
// Intro
// When writing shader or during any procedural creation process
//  (texturing, modeling, shading, animation...) you often find yourself
//  modifying signals in different ways so they behave the way you need.
// It is common to use smoothstep() to threshold some values, or pow() to
//  shape a signal, or clamp() to clip it, fmod() to make it repeat, a mix()
//  to blend between two signals, exp() for attenuation, etc etc.
// All these functions are often conveniently available by default in most languages.
// However there are some operations that are also relatively used that don't come
//  by default in any language.
//  The following is a list of some of the functions that I find myself using over and over again:

// Almost Identity (I)
// Imagine you don't want to modify a signal unless it's drops to zero or close to it,
//  in which case you want to replace the value with a small positive constant.
// Then, rather than clamping the value and introduce a discontinuity,
//  you can smoothly blend the signal into the desired clipped value.
// So, let m be the threshold (anything above m stays unchanged),
//  and n the value things will take when the signal is zero.
// Then, the following function does the soft clipping (in a cubic fashion):
float almostIdentity1(float x, float m, float n) {
    if (x > m) return x;
    float a = 2.0*n - m;
    float b = 2.0*m - 3.0*n;
    float t = x/m;
    return (a*t + b)*t*t + n;
}

// Almost Identity (II)
// A different way to achieve a near identity is through the square root of a biased square.
// I saw this technique first in a shader by user "omeometo" in Shadertoy.
// This approach can be a bit slower than the cubic above, depending on the hardware,
//  but I find myself using it a lot these days. While it has zero derivative,
//  it has a non-zero second derivative, so keep an eye in case it causes problems in your application.
// An extra nice thing is that this function can be used, unaltered, as an smooth-abs() function,
//  which comes handy for symmetric funtions such as mirrored SDFs.
float almostIdentity2(float x, float n) {
  return sqrt(x*x+n);
}

// Almost Unit Identity
// This is a near-identiy function that maps the unit interval into itself.
// It is the cousin of smoothstep(), in that it maps 0 to 0, 1 to 1, 
//  and has a 0 derivative at the origin, just like smoothstep. 
// However, instead of having a 0 derivative at 1, it has a derivative of 1 at that point.
// It's equivalent to the Almost Identiy above with n=0 and m=1.
// Since it's a cubic just like smoothstep() it is very fast to evaluate.
float almostUnitIdentity(float x) {
  return x*x*(2.0-x);
}

// If you use smoothstep for a velocity signal
//  (say, you want to smoothly accelerate a stationary object into constant velocity motion),
//  you need to integrate smoothstep() over time in order to get the actual position of
//  value of the animation.
// The function below is exactly that, the position of an object that accelerates with smoothstep.
// Note it's derivative is never larger than 1, so no decelerations happen.
float integralSmoothstep(float x, float T) {
  if(x > T) return x - T/2.0;
  return x*x*x*(1.0-x*0.5/T)/T/T;
}

// Exponential Impulse
// Impulses are great for triggering behaviours or making envelopes for music or animation.
// Basically, for anything that grows fast and then decays slowly.
// The following is an exponential impulse function.
// Use k to control the stretching of the function.
// Its maximum, which is 1, happens at exactly x = 1/k.
float expImpulse(float x, float k) {
  float h = k*x;
  return h*exp(1.0-h);
}

// Polynomial Impulse
// Another impulse function that doesn't use exponentials can be designed by using polynomials.
// Use k to control falloff of the function.
// For example, a quadratic can be used, which peaks at x = sqrt(1/k).
float quaImpulse(float k, float x) {
  return 2.0*sqrt(k)*x/(1.0+k*x*x);
}

// You can easily generalize it to other powers to get different falloff shapes,
//  where n is the degree of the polynomial.
// These generalized impulses peak at x = [k(n-1)]-1/n.
float polyImpulse(float k, float n, float x) {
    return (n/(n-1.0))*pow((n-1.0)*k,1.0/n)*x/(1.0+k*pow(x,n));
}

// Sustained Impulse
// Similar to the previous, but it allows for control on the width of attack
//  (through the parameter "k") and the release (parameter "f") independently. 
// Also, the impulse releases at a value of 1 instead of 0.
float expSustainedImpulse(float x, float f, float k) {
  float s = max(x-f,0.0);
  return min(x*x/(f*f), 1+(2.0/f)*s*exp(-k*s));
}

// Cubic Pulse
// Chances are you found yourself doing smoothstep(c-w,c,x)-smoothstep(c,c+w,x) very often.
// I do, for example when I need to isolate some features in a signal. 
// For those cases, this cubicPulse() below is my new friend and will be yours too soon.
// Bonus - you can also use it as a performant replacement for a gaussian.
float cubicPulse(float c, float w, float x) {
  x = abs(x - c);
  if(x > w) return 0.0;
  x /= w;
  return 1.0 - x*x*(3.0-2.0*x);
}

// Exponential Step
// A natural attenuation is an exponential of a linearly decaying quantity: yellow curve, exp(-x). 
// A gaussian, is an exponential of a quadratically decaying quantity: light green curve, exp(-x2). 
// You can generalize and keep increasing powers, and get a sharper and sharper s-shaped curves.
// For really high values of n you can approximate a perfect step().
// If you want such step to transition at x=a, like in the graphs to the right, 
//  you can set k = a-n⋅ln 2.
float expStep(float x, float k, float n) {
  return exp(-k*pow(x,n));
}

// Gain
// Remapping the unit interval into the unit interval by expanding the sides and 
//  compressing the center, and keeping 1/2 mapped to 1/2, that can be done with the gain() function. 
// This was a common function in RSL tutorials (the Renderman Shading Language). 
// k=1 is the identity curve, k<1 produces the classic gain() shape, and k>1 produces "s" shaped curves
// The curves are symmetric (and inverse) for k=a and k=1/a.
float gain(float x, float k) {
  float a = 0.5*pow(2.0*((x<0.5)?x:1.0-x), k);
  return (x<0.5)?a:1.0-a;
}

// Parabola
// A nice choice to remap the 0..1 interval into 0..1, 
//  such that the corners are mapped to 0 and the center to 1. 
// You can then rise the parabola to a power k to control its shape.
float parabola(float x, float k) {
  return pow(4.0*x*(1.0-x), k);
}

// Power curve
// This is a generalization of the Parabola() above. 
// It also maps the 0..1 interval into 0..1 by keeping the corners mapped to 0. 
// But in this generalization you can control the shape one either side of the curve, 
//  which comes handy when creating leaves, eyes, and many other interesting shapes.
// Note that k is chosen such that pcurve() reaches exactly 1 at its maximum for illustration purposes,
//  but in many applications the curve needs to be scaled anyways so the 
//  slow computation of k can be simply avoided.
float pcurve(float x, float a, float b) {
  float k = pow(a+b,a+b)/(pow(a,a)*pow(b,b));
  return k*pow(x,a)*pow(1.0-x,b);
}

// Sinc curve
// A phase shifted sinc curve can be useful if it starts at zero and ends at zero, 
//  for some bouncing behaviors (suggested by Hubert-Jan). 
// Give k different integer values to tweak the amount of bounces. 
// It peaks at 1.0, but that take negative values, which can make it unusable in some applications.
float sinc(float x, float k) {
  float a = PI*((k*x-1.0));
  return sin(a)/a;
}

// color in lines
vec3 color_fg(float pct, vec3 current, vec3 color) {
  // return (1.0 - step(pct, 0.9)) * color + step(pct, 0.1) * current;
  return (1.0 - pct) * current + pct * color;
}

// plot a line on Y using a value between 0.0-1.0
float plot(vec2 st, float pct) {
  // aye look at that better way to plot thanks to this
  // aye does make a good gaussian blur for large w in this
  return cubicPulse(pct, 0.010, st.y);
  // return 
  //   smoothstep(
  //     0.2, 1.0,
  //     smoothstep(pct-0.1, pct, st.y) 
  //     - smoothstep(pct, pct+0.01, st.y)
  //   );
}

void main() {
  // vec2 st = gl_FragCoord.xy/u_resolution;
  vec2 st = fragTexCoord.xy;

  // bg color
  vec3 color = vec3(0.0);


  // speed
  float speed = u_time * 0.5;

  // movement along x
  float x = st.x; // none
  // float x = sin(speed + st.x * PI * 2) / 2 - 0.5; // focus center
  // float x = sin(speed + st.x*PI); // full

  // movement of exp
  float e = 1; // linear
  // float e = map(sin(2*speed), -1.0, 1.0, 0.5, 3.5); // move exp

  // movement along 01
  // float zo = 2;
  // float zoo = 3;
  // float zoo = 3;
  float zo = map(sin(speed), -1.0, 1.0, 0.0, 1.0);
  float zoo = map(cos(2.0*speed), -1.0, 1.0, 0.0, 1.0);
  float nzoo = map(cos(2.0*speed), -1.0, 1.0, 0.001, 1.0);

  // Almost Identity (I)
  if (1 == 1)
  color = color_fg(
    plot(st, 
      almostIdentity1(1.0 - pow(x, e), zoo, zo)
    ),
    color, vec3(0.0, 1.0, 0.0)
  );

  // Almost Identity (II)
  if (1 == 1)
  color = color_fg(
    plot(st, 
      almostIdentity2(1.0 - pow(x, e), zo)
    ),
    color, vec3(1.0, 0.0, 0.0)
  );

  // Almost Unit Identity
  if (1 == 1)
  color = color_fg(
    plot(st, 
      almostUnitIdentity(1.0 - pow(x, e))
    ),
    color, vec3(0.0, 0.0, 1.0)
  );

  // Integral Smoothstep
  if (1 == 1)
  color = color_fg(
    plot(st, 
      integralSmoothstep(1.0 - pow(x, e), zo)
    ),
    color, vec3(1.0, 0.0, 1.0)
  );

  // Exponential Impulse
  if (1 == 1)
  color = color_fg(
    plot(st, 
      expImpulse(1.0 - pow(x, e), zo)
    ),
    color, vec3(0.0, 1.0, 1.0)
  );

  // Quadratic Impulse
  if (1 == 1)
  color = color_fg(
    plot(st, 
      quaImpulse(1.0 - pow(x, e), zo)
    ),
    color, vec3(1.0, 1.0, 0.0)
  );

  // Poly Impulse
  if (1 == 1)
  color = color_fg(
    plot(st, 
      polyImpulse(1.0 - pow(x, e), nzoo * 8, zo)
    ),
    color, vec3(1.0, 1.0, 1.0)
  );

  // Sustained Impulse
  if (1 == 1)
  color = color_fg(
    plot(st, 
      expSustainedImpulse(1.0 - pow(x, e), zoo * 5, zo * 2)
    ),
    color, vec3(1.0, 0.5, 1.0)
  );

  // Cubic Pulse 
  if (1 == 1)
  color = color_fg(
    plot(st, 
      cubicPulse(zoo * 3, zo * 2, 1.0 - pow(x, e))
    ),
    color, vec3(0.8, 0.5, 0.4)
  );

  // Exponential Step
  if (1 == 1)
  color = color_fg(
    plot(st, 
      expStep(1.0 - pow(x, e), zoo * 9, zo * 9)
    ),
    color, vec3(0.0, 1.0, 0.5)
  );

  // Gain
  if (1 == 1)
  color = color_fg(
    plot(st, 
      gain(1.0 - pow(x, e), zo * 3)
    ),
    color, vec3(0.5, 0.5, 0.0)
  );

  // Parabola
  if (1 == 1)
  color = color_fg(
    plot(st, 
      parabola(1.0 - pow(x, e), zo * 9)
    ),
    color, vec3(0.5, 0.5, 0.7)
  );

  // Power Curve
  if (1 == 1)
  color = color_fg(
    plot(st, 
      pcurve(1.0 - pow(x, e), zo * 3, zoo * 8)
    ),
    color, vec3(0.1, 0.2, 0.7)
  );

  // Sinc Curve
  if (1 == 1)
  color = color_fg(
    plot(st, 
      sinc(1.0 - pow(x, e), zo * 13) + 0.5
    ),
    color, vec3(1.0, 0.2, 0.7)
  );

	outputColor = vec4(color, 1.0);
}
