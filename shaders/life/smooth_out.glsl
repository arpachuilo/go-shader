#version 410
uniform sampler2D inputA;
uniform vec2 scale;

in vec2 fragTexCoord;

out vec4 outputColor;

void main()
{
    vec2 uv = fragTexCoord.xy;
    vec4 col = texture(inputA, uv);
    
    outputColor = col.x*vec4(1.0) + col.y*vec4(1,0.5,0,0) + col.z*vec4(0,0.5,1,0);
}
