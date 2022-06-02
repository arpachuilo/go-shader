#version 410
uniform int maxIterations;
uniform vec2 focus;
uniform vec2 offset;
uniform float zoom;

uniform vec2 scale;

in vec2 fragTexCoord;
out vec4 outputColor;

void main() {      
  vec2 c = vec2(gl_FragCoord.xy);
  /* c = (c * exp(-zoom)); */

  int iteration = 0;
  float zx = 1.5 * (c.x - scale.x / 2) / (0.5 * scale.x);
  float zy = (c.y - scale.y/2) / (0.5 * scale.y);

  zx = (zx * exp(-zoom)) + offset.x;
  zy = (zy * exp(-zoom)) + offset.y;
  while(zx*zx + zy*zy < 4.0 && iteration < maxIterations) {
      float tmp = zx*zx - zy*zy + focus.x;
      zy = 2.0*zx*zy + focus.y;
      zx = tmp;

      // Complex multiplication, then addition
      ++iteration;
  }


  // Generate the colors
  outputColor = vec4(float(iteration) / float(maxIterations));
}
