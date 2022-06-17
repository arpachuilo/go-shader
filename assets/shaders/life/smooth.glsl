#version 410
uniform sampler2D inputA;
uniform sampler2D inputC;

uniform float cursorSize;
uniform float time;
uniform vec2 scale;
uniform vec4 mouse;

uniform float or = 18.0;         // outer gaussian std dev
uniform float ir = 6.0;          // inner gaussian std dev

const float b1 = 0.19;         // birth1
const float b2 = 0.212;        // birth2
const float s1 = 0.267;        // survival1
const float s2 = 0.445;        // survival2
const float dt = 0.2;          // timestep
const float alpha_n = 0.017;   // sigmoid width for outer fullness
const float alpha_m = 0.112;   // sigmoid width for inner fullness

in vec2 fragTexCoord;

out vec4 outputColor;

// the logistic function is used as a smooth step function
float sigma1(float x,float a,float alpha) 
{ 
    return 1.0 / (1.0 + exp(-(x-a)*4.0/alpha));
}

float sigma2(float x,float a,float b,float alpha)
{
    return sigma1(x,a,alpha) 
        * (1.0-sigma1(x,b,alpha));
}

float sigma_m(float x,float y,float m,float alpha)
{
    return x * (1.0-sigma1(m,0.5,alpha)) 
         + y * sigma1(m,0.5,alpha);
}

// the transition function
// (n = outer fullness, m = inner fullness)
float s(float n,float m)
{
    return sigma2(n, sigma_m(b1,s1,m,alpha_m), 
        sigma_m(b2,s2,m,alpha_m), alpha_n);
}

void main() {
    vec2 tx = 1.0 / scale.xy;
    vec2 uv = gl_FragCoord.xy * tx;
	
    const float _K0 = -20.0/6.0; // center weight
    const float _K1 = 4.0/6.0;   // edge-neighbors
    const float _K2 = 1.0/6.0;   // vertex-neighbors
    
    vec4 current = texture(inputA, uv);
    vec2 fullness = texture(inputC, uv).xy;
    
    float delta =  2.0 * s(fullness.x, fullness.y) - 1.0;
    float new = clamp(current.x + dt * delta, 0.0, 1.0);
    
    float mouse_distance = length(gl_FragCoord.xy - mouse.xy);
    if (mouse.z > 0.0) {
        // from chronos' SmoothLife shader https://www.shadertoy.com/view/XtdSDn
        if (mouse_distance <= or) {
        	new = step((ir+1.5), mouse_distance) * (1.0 - step(or, mouse_distance));
        }
    }
    
    outputColor = vec4(new, fullness, current.w);
}
