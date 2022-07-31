// sample state using uv and apply magma coloring
#version 410
uniform int index;
uniform sampler2D state;
uniform vec2 scale;
uniform float alpha;

vec3 magma(float t) {
  const vec3 c0 = vec3(-0.002136485053939582, -0.000749655052795221, -0.005386127855323933);
  const vec3 c1 = vec3(0.2516605407371642, 0.6775232436837668, 2.494026599312351);
  const vec3 c2 = vec3(8.353717279216625, -3.577719514958484, 0.3144679030132573);
  const vec3 c3 = vec3(-27.66873308576866, 14.26473078096533, -13.64921318813922);
  const vec3 c4 = vec3(52.17613981234068, -27.94360607168351, 12.94416944238394);
  const vec3 c5 = vec3(-50.76852536473588, 29.04658282127291, 4.23415299384598);
  const vec3 c6 = vec3(18.65570506591883, -11.48977351997711, -5.601961508734096);

  return c0+t*(c1+t*(c2+t*(c3+t*(c4+t*(c5+t*c6)))));
}

in vec2 fragTexCoord;

out vec4 outputColor;

void main() {
  vec4 tex = texture(state, gl_FragCoord.xy  / scale, 0);
  vec4 color = vec4(magma(tex[index]), 1.0);
  outputColor = vec4(color.rgb, 1.0 - color.a * alpha);
}
