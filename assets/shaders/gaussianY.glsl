#version 410
#define SQRT_2_PI 2.50662827463

uniform sampler2D inputB;

uniform int frame;
uniform float cursorSize;
uniform float time;
uniform vec2 scale;
uniform vec2 mouse;

// ---------------------------------------------
uniform float or = 18.0;         // outer gaussian std dev
uniform float ir = 6.0;          // inner gaussian std dev
const int   oc = 50;           // sample cutoff
// ---------------------------------------------


in vec2 fragTexCoord;

out vec4 outputColor;


vec2 gaussian(float i, vec2 a, vec2 d) {
     return a * exp( -(i*i) / d );
}

vec2 gaussian1d(sampler2D sam, vec2 sigma, vec2 uv, vec2 tx) {
    vec2 a = vec2(1.0 / (sigma * SQRT_2_PI));
    vec2 d = vec2(2.0 * sigma * sigma);
    vec2 acc = vec2(0.0);
    vec2 sum = vec2(0.0);
    
    // centermost term
    acc += a * texture(sam, uv).x;
    sum += a;

    // sum up remaining terms symmetrically
    for (int i = 1; i <= oc; i++) {
        float fi = float(i);
        vec2 g = gaussian(fi, a, d);
        vec2 posL = fract(uv - tx * fi);
        vec2 posR = fract(uv + tx * fi);
        acc += g * (texture(sam, posL).xy + texture(sam, posR).xy);
        sum += 2.0 * g;
    }

    return acc / sum;
}

void main()
{
    vec2 tx = 1.0 / scale.xy;
    vec2 uv = gl_FragCoord.xy * tx;
    tx = (mod(float(frame),2.0) < 1.0) ? vec2(0,tx.y) : vec2(tx.x,0);
    vec2 y_pass = gaussian1d(inputB, vec2(or, ir), uv, tx);
    outputColor = vec4(y_pass,0,0);
}
