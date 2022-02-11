#version 410
uniform sampler2D state;
uniform vec2 scale;

in vec2 fragTexCoord;
out vec4 outputColor;

void main() {
    vec4 tex = texture(state, fragTexCoord.xy);
    outputColor = tex;
}
