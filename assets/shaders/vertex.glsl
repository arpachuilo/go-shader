#version 410

in vec2 vert;
in vec2 vertTexCoord;

out vec2 fragTexCoord;

void main() {
    gl_Position = vec4(vert,0,1);
    fragTexCoord = vertTexCoord;
}
