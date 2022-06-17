#version 410
uniform int index;
uniform sampler2D state;
uniform vec2 scale;

const float pi = 3.141592653589793238462643383;
const float pi1_3 = pi / 3;
const float pi2_3 = pi * 2 / 3;

vec3 sinebow(float t) {
    t = (0.5 - t) * pi;

    return vec3(
      pow(sin(t), 2),
      pow(sin(t+pi1_3), 2),
      pow(sin(t+pi2_3), 2)
    );
}

in vec2 fragTexCoord;

out vec4 outputColor;

void main() {
    vec4 tex = texture(state, fragTexCoord.xy);
    vec4 color = vec4(sinebow(tex[index]), 1.0);
    outputColor = vec4(color.rgb, 1.0);
}
