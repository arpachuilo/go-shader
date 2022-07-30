#version 410
uniform int index;
uniform sampler2D state;
uniform vec2 scale;
uniform float alpha;

vec3 cividis(float t) {
    float r = round(-4.54 - t*(35.34-t*(2381.73-t*(6402.7-t*(7024.72-t*2710.57)))));
    float g = round(32.49 + t*(170.73+t*(52.82-t*(131.46-t*(176.58-t*67.37)))));
    float b = round(81.24 + t*(442.36-t*(2482.43-t*(6167.24-t*(6614.94-t*2475.67)))));

    return vec3(
      r / 255, 
      g / 255, 
      b / 255
    );
}

in vec2 fragTexCoord;

out vec4 outputColor;

void main() {
    vec4 tex = texture(state, fragTexCoord.xy);
    vec4 color = vec4(cividis(tex[index]), 1.0);
    outputColor = vec4(color.rgb, 1.0 - color.a * alpha);
}
