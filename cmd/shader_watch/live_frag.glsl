#version 410

// https://learnopengl.com/Lighting/Colors
uniform vec2 u_resolution;
uniform vec2 u_mouse;
uniform float u_time;
uniform int u_frame;
uniform float u_farclip;

// matrices
uniform mat4 ModelMatrix;
uniform mat4 ViewMatrix;
uniform mat4 ProjectionMatrix;

// mat
struct Material {
    vec3 ambient;
    vec3 diffuse;
    vec3 specular;
    float shininess;
}; 
  
uniform Material material;

// lights
struct Light {
    vec3 position;
  
    vec3 ambient;
    vec3 diffuse;
    vec3 specular;
};

uniform Light light; 

in vec4 ex_wposition;
in vec4 ex_position;

in vec2 ex_tex;

in vec3 ex_normal;
in vec3 ex_wnormal;

layout(location = 0) out vec4 outputColor;
layout(location = 1) out vec4 outputNormal;

vec3 turbo(float x) {
  float r = 0.1357 + x * ( 4.5974 - x * ( 42.3277 - x * ( 130.5887 - x * ( 150.5666 - x * 58.1375 ))));
  float g = 0.0914 + x * ( 2.1856 + x * ( 4.8052 - x * ( 14.0195 - x * ( 4.2109 + x * 2.7747 ))));
  float b = 0.1067 + x * ( 12.5925 - x * ( 60.1097 - x * ( 109.0745 - x * ( 88.5066 - x * 26.8183 ))));
  return vec3(r,g,b);
}

// phong https://learnopengl.com/Lighting/Basic-Lighting
void main() {
  // vec3 tColor = turbo(1.0 - ex_wposition.w/u_farclip*2);

  // ambient
  vec3 ambient = light.ambient * material.ambient;

  // diffuse
  vec3 norm = normalize(ex_normal);
  vec3 lightDir = normalize(light.position - ex_position.xyz);
  float diff = max(dot(norm, lightDir), 0.0);
  vec3 diffuse = light.diffuse * (diff * material.diffuse);

  // specular
  vec3 viewDir = normalize(ViewMatrix[3].xyz - ex_position.xyz);
  vec3 reflectDir = reflect(-lightDir, norm);
  float spec = pow(max(dot(viewDir, reflectDir), 0.0), material.shininess);
  vec3 specular = light.specular * (spec * material.specular);

  // combine
  vec3 color = (ambient + diffuse + specular);

  // output
  // color = vec3(1.0);
  outputColor = vec4(color, 1.0);
  outputNormal = vec4(ex_normal, 1.0);
}

