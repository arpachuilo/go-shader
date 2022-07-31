// sample state using uv and apply plasma coloring
#version 410
uniform int index;
uniform sampler2D state;
uniform vec2 scale;
uniform float alpha;


vec3 plasma(float t) {
  const vec3 c0 = vec3(0.05873234392399702, 0.02333670892565664, 0.5433401826748754);
  const vec3 c1 = vec3(2.176514634195958, 0.2383834171260182, 0.7539604599784036);
  const vec3 c2 = vec3(-2.689460476458034, -7.455851135738909, 3.110799939717086);
  const vec3 c3 = vec3(6.130348345893603, 42.3461881477227, -28.51885465332158);
  const vec3 c4 = vec3(-11.10743619062271, -82.66631109428045, 60.13984767418263);
  const vec3 c5 = vec3(10.02306557647065, 71.41361770095349, -54.07218655560067);
  const vec3 c6 = vec3(-3.658713842777788, -22.93153465461149, 18.19190778539828);

  return c0+t*(c1+t*(c2+t*(c3+t*(c4+t*(c5+t*c6)))));
}

in vec2 fragTexCoord;

out vec4 outputColor;

void main() {
  vec4 tex = texture(state, gl_FragCoord.xy  / scale, 0);
  vec4 color = vec4(plasma(tex[index]), 1.0);
  outputColor = vec4(color.rgb, 1.0 - color.a * alpha);
}
