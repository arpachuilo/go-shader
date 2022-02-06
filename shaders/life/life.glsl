#version 410
uniform int s[9];
uniform int b[9];
uniform sampler2D state;
uniform vec2 scale;

in vec2 fragTexCoord;

out vec4 outputColor;

vec4 get(vec2 coord) {
    return texture(state, vec2(gl_FragCoord.xy + coord) / scale, 0);
}

ivec4 alive(vec4 cell) {
	return ivec4(
	  cell.r > 0.0 ? 1 : 0,
	  cell.g > 0.0 ? 1 : 0,
	  cell.b > 0.0 ? 1 : 0,
	  cell.a > 0.0 ? 1 : 0
	);
}

float op(float c, int n) {
    if (n == b[0] || n == b[1] || n == b[2] || n == b[3] || n == b[4] || n == b[5] || n == b[6] || n == b[7] || n == b[8]) {
    	return 1.0;
    } else if (n == s[0] || n == s[1] || n == s[2] || n == s[3] || n == s[4] || n == s[5] || n == s[6] || n == s[7] || n == s[8]) {
      	return c;
    }

    return 0.0;
}

void main() {
    // sum each channel alive
    ivec4 sum = alive(get(vec2(-1, -1))) +
               alive(get(vec2(-1,  0))) +
               alive(get(vec2(-1,  1))) +
               alive(get(vec2( 0, -1))) +
               alive(get(vec2( 0,  1))) +
               alive(get(vec2( 1, -1))) +
               alive(get(vec2( 1,  0))) +
               alive(get(vec2( 1,  1)));

    vec4 current = get(vec2(0, 0));
    outputColor = vec4(
    	op(current.r, sum.r),
    	op(current.g, sum.g),
    	op(current.b, sum.b),
    	op(current.a, sum.a)
    );
}
