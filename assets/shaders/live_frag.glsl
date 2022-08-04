#version 410

uniform vec2 u_resolution;
uniform vec2 u_mouse;
uniform float u_time;
uniform int u_frame;
uniform float u_farclip;

in vec4 fragTexCoord;
in vec4 ex_position;

out vec4 outputColor;

vec3 turbo(float x) {
  float r = 0.1357 + x * ( 4.5974 - x * ( 42.3277 - x * ( 130.5887 - x * ( 150.5666 - x * 58.1375 ))));
  float g = 0.0914 + x * ( 2.1856 + x * ( 4.8052 - x * ( 14.0195 - x * ( 4.2109 + x * 2.7747 ))));
  float b = 0.1067 + x * ( 12.5925 - x * ( 60.1097 - x * ( 109.0745 - x * ( 88.5066 - x * 26.8183 ))));
  return vec3(r,g,b);
}

// float getGrayScale(Vec4 color, vec2 coods){
// 	vec4 color = texture(sampler, coods);
// 	float gray = (color.r + color.g + color.b)/3.0;
// 	return gray;
// }

void main() {
 //  vec2 delta = vec2(0.0,0.003);
	// vec2 iResolution = 1.0 / u_resolution;
	// float m = max(iResolution.x,iResolution.y);
	// vec2 texCoords = fragTexCoord;
	// 
 //  vec3 screen_color = gl_FragColor
	// vec3 screen_color = texture(SCREEN_TEXTURE, SCREEN_UV).rgb;
	// 
	// float c1y = getGrayScale(SCREEN_TEXTURE, texCoords.xy-delta/2.0);
	// float c2y = getGrayScale(SCREEN_TEXTURE, texCoords.xy+delta/2.0);
	// 
	// float c1x = getGrayScale(SCREEN_TEXTURE, texCoords.xy-delta.yx/2.0);
	// float c2x = getGrayScale(SCREEN_TEXTURE, texCoords.xy+delta.yx/2.0);
	// 
	// float dcdx = (c2x - c1x)/(delta.y*10.0);
	// float dcdy = (c2y - c1y)/(delta.y*10.0);
	// 
	// vec2 dcdi = vec2(dcdx,dcdy);
	// float edge = length(dcdi)/10.0;
	// edge = 1.0 - edge;
	// edge = smoothstep(threshold, threshold + blend, edge);
	// 
	// COLOR.rgb = Vec4(mix(edge_color.rgb, screen_color.rgb, edge), 1.0);
  outputColor = vec4(turbo(1.0 - ex_position.w/u_farclip), 1.0);
}
