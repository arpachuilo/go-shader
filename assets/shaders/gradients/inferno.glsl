// sample state using uv and apply inferno coloring
#version 410
uniform int index;
uniform sampler2D state;
uniform vec2 scale;
uniform float alpha;

vec3 inferno(float t) {
  const vec3 c0 = vec3(0.0002189403691192265, 0.001651004631001012, -0.01948089843709184);
  const vec3 c1 = vec3(0.1065134194856116, 0.5639564367884091, 3.932712388889277);
  const vec3 c2 = vec3(11.60249308247187, -3.972853965665698, -15.9423941062914);
  const vec3 c3 = vec3(-41.70399613139459, 17.43639888205313, 44.35414519872813);
  const vec3 c4 = vec3(77.162935699427, -33.40235894210092, -81.80730925738993);
  const vec3 c5 = vec3(-71.31942824499214, 32.62606426397723, 73.20951985803202);
  const vec3 c6 = vec3(25.13112622477341, -12.24266895238567, -23.07032500287172);

  return c0+t*(c1+t*(c2+t*(c3+t*(c4+t*(c5+t*c6)))));
}

in vec2 fragTexCoord;

out vec4 outputColor;

void main() {
  vec4 tex = texture(state, gl_FragCoord.xy  / scale, 0);
  vec4 color = vec4(inferno(tex[index]), 1.0);
  outputColor = vec4(color.rgb, 1.0 - color.a * alpha);
}
