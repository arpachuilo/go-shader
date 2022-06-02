#version 410
uniform int maxIterations;
uniform vec2 focus;
uniform float zoom;

uniform vec2 scale;

in vec2 fragTexCoord;
out vec4 outputColor;

void main() {      
  // C is the aspect-ratio corrected UV coordinate.
  vec2 c = (-1.0 + 2.0 * gl_FragCoord.xy / scale.xy) * vec2(scale.x / scale.y, 1.0);
  c = (c * exp(-zoom)) + focus;


  /* vec2 c = gl_FragCoord.xy / abs(scale/zoom) + focus; */
  /* c = (c * exp(-zoom)) + focus; */

  vec2 z = c;
  int iteration = 0;

  while(iteration < maxIterations) {
      // Precompute for efficiency
      float zr2 = z.x * z.x;
      float zi2 = z.y * z.y;

      // The larger the square length of Z,
      // the smoother the shading
      if(zr2 + zi2 > 128.0) break;

      // Complex multiplication, then addition
      z = vec2(zr2 - zi2, 2.0 * z.x * z.y) + c;
      ++iteration;
  }


  // Generate the colors
  outputColor = vec4(float(iteration) / float(maxIterations));
}
