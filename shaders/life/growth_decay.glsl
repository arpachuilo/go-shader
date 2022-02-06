#version 410
uniform sampler2D state;
uniform sampler2D self;
uniform vec2 scale;

in vec2 fragTexCoord;
out vec4 outputColor;

const float gain = 0.3;
const float decay = -0.01;

vec4 getCell(vec2 coord) {
    return texture(state, vec2(gl_FragCoord.xy + coord) / scale, 0);
}

vec4 getSelf(vec2 coord) {
    return texture(self, vec2(gl_FragCoord.xy + coord) / scale, 0);
}

float update(float cell, float current) {
	float offset = cell > 0.0 ? gain : decay;
	return current + offset;
}

void main() {
    vec4 cell = getCell(vec2(0, 0));
    vec4 self = getSelf(vec2(0, 0));
    outputColor = vec4(
    	update(cell.r, self.r),
    	update(cell.g, self.g),
    	update(cell.b, self.b),
    	update(cell.a, self.a)
    );
}
