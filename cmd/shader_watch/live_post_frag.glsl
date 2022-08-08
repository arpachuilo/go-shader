#version 410

uniform vec2 u_resolution;
uniform vec2 u_mouse;
uniform float u_time;
uniform int u_frame;

uniform sampler2D color_buffer;
uniform sampler2D normal_buffer;

in vec2 fragTexCoord;
in vec4 ex_position;

layout(location = 0) out vec4 outputColor;

vec4 get(sampler2D sampler, vec2 coord) {
  return texture(sampler, vec2(gl_FragCoord.xy + coord) / u_resolution, 0);
}

vec4 grey(vec4 c) {
  float average = 0.2126 * c.r + 0.7152 * c.g + 0.0722 * c.b;
  return vec4(average, average, average, 1.0);
}

// https://learnopengl.com/Advanced-OpenGL/Framebuffers
void main() {
  vec4 current = get(color_buffer, vec2(0, 0));
  // vec4 current = get(normal_buffer, vec2(0, 0));

  // no change
  // outputColor = current;

  // invert
  // outputColor = vec4(vec3(1.0 - current).xyz, 1.0);

  // grey scale
  // outputColor = grey(current);

  // kernel effects
  float offset = 1.0f;
  vec2 offsets[9] = vec2[](
      vec2(-offset,  offset), // top-left
      vec2( 0.0f,    offset), // top-center
      vec2( offset,  offset), // top-right
      vec2(-offset,  0.0f),   // center-left
      vec2( 0.0f,    0.0f),   // center-center
      vec2( offset,  0.0f),   // center-right
      vec2(-offset, -offset), // bottom-left
      vec2( 0.0f,   -offset), // bottom-center
      vec2( offset, -offset)  // bottom-right    
  );

  // sharpen
  // float kernel[9] = float[](
  //     -1, -1, -1,
  //     -1,  9, -1,
  //     -1, -1, -1
  // );

  // blur
  // float kernel[9] = float[](
  //     1.0/16, 2.0/16, 1.0/16,
  //     2.0/16, 4.0/16, 2.0/16,
  //     1.0/16, 2.0/16, 1.0/16
  // );

  // edge detection
  // float kernel[9] = float[](
  //     1, 1, 1,
  //     1, -8, 1,
  //     1, 1, 1
  // );

  // vec4 sampleTex[9];
  // for(int i = 0; i < 9; i++) {
  //   sampleTex[i] = get(offsets[i]);
  // }
  //
  // vec4 col = vec4(0.0);
  // for(int i = 0; i < 9; i++) {
  //   col += sampleTex[i] * kernel[i];
  // }

  // prewittx
  // float prewittx[9] = float[](
  //     1, 0, -1,
  //     2, 0, -2,
  //     1, 0, -1
  // );

  // prewitty
  // float prewitty[9] = float[](
  //     1, 2, 1,
  //     0, 0, 0,
  //    -1,-2,-1
  // );


  // sobelx
  float sobelx[9] = float[](
      1, 0, -1,
      2, 0, -2,
      1, 0, -1
  );

  // sobely
  float sobely[9] = float[](
      1, 2, 1,
      0, 0, 0,
     -1,-2,-1
  );

  vec4 sampleTex[9];
  for(int i = 0; i < 9; i++) {
    sampleTex[i] = get(normal_buffer, offsets[i]);
    // sampleTex[i] = get(color_buffer, offsets[i]);
  }

  vec4 colX = vec4(0.0);
  vec4 colY = vec4(0.0);
  for(int i = 0; i < 9; i++) {
    colX += sampleTex[i] * sobelx[i];
    colX += sampleTex[i] * sobelx[i];
    // colY += sampleTex[i] * prewitty[i];
    // colY += sampleTex[i] * prewitty[i];
  }

  vec4 col = sqrt(colX*colX + colY*colY);
  // vec4 col = vec4(f, f, f, 1.0);

  // luminance
  float l = 0.299*col.r + 0.587*col.g + 0.114*col.b;

  // norm
  // float l = length(col);
  if (l > 0.1) {
    outputColor = vec4(1.0);
  } else {
    outputColor = current;
  }
}
