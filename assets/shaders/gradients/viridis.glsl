// sample state using uv and apply viridis coloring
#version 410
uniform int index;
uniform sampler2D state;
uniform vec2 scale;
uniform float alpha;

vec3 viridis(float t) {
  const vec3 c0 = vec3(0.2777273272234177, 0.005407344544966578, 0.3340998053353061);
  const vec3 c1 = vec3(0.1050930431085774, 1.404613529898575, 1.384590162594685);
  const vec3 c2 = vec3(-0.3308618287255563, 0.214847559468213, 0.09509516302823659);
  const vec3 c3 = vec3(-4.634230498983486, -5.799100973351585, -19.33244095627987);
  const vec3 c4 = vec3(6.228269936347081, 14.17993336680509, 56.69055260068105);
  const vec3 c5 = vec3(4.776384997670288, -13.74514537774601, -65.35303263337234);
  const vec3 c6 = vec3(-5.435455855934631, 4.645852612178535, 26.3124352495832);

  return c0+t*(c1+t*(c2+t*(c3+t*(c4+t*(c5+t*c6)))));
}

in vec2 fragTexCoord;

out vec4 outputColor;

void main() {
    vec4 tex = texture(state, gl_FragCoord.xy  / scale, 0);
    vec4 color = vec4(viridis(tex[index]), 1.0);
    outputColor = vec4(color.rgb, 1.0 - color.a * alpha);
}
