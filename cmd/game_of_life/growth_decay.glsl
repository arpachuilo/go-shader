#version 410
uniform sampler2D state;
uniform sampler2D self;
uniform vec2 u_resolution;

in vec2 fragTexCoord;
out vec4 outputColor;

const float gain = 0.3;
const float decay = -0.01;

float update(float cell, float current) {
	float offset = cell > 0.5 ? gain : decay;
	return current + offset;
}

void main() {
  vec4 cell = texture(state, gl_FragCoord.xy/u_resolution, 0);
  vec4 self = texture(self, gl_FragCoord.xy/u_resolution, 0);
  outputColor = vec4(
    update(cell.r, self.r),
    update(cell.g, self.g),
    update(cell.b, self.b),
    update(cell.a, self.a)
  );
}
