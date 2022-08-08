#version 410

uniform vec2 u_resolution;
uniform vec2 u_mouse;
uniform float u_time;
uniform int u_frame;

uniform mat4 ModelMatrix;
uniform mat4 ViewMatrix;
uniform mat4 ProjectionMatrix;

layout(location = 0) in vec3 pos;
layout(location = 1) in vec2 tex;
layout(location = 2) in vec3 normal;

out vec4 ex_wposition;
out vec4 ex_position;

out vec2 ex_tex;

out vec3 ex_wnormal;
out vec3 ex_normal;

void main() {
  gl_Position = (ProjectionMatrix * ViewMatrix * ModelMatrix) * vec4(pos, 1.0);
  ex_wposition = gl_Position;
  ex_position = ModelMatrix * vec4(pos, 1.0);

  ex_tex = tex;

  ex_wnormal = normal;
  ex_normal = mat3(transpose(inverse(ModelMatrix))) * normal;  
  // from https://learnopengl.com/Lighting/Basic-Lighting
  // Inversing matrices is a costly operation for shaders, so wherever possible try to avoid doing inverse operations since they have to be done on each vertex of your scene. For learning purposes this is fine, but for an efficient application you'll likely want to calculate the normal matrix on the CPU and send it to the shaders via a uniform before drawing (just like the model matrix).
}
