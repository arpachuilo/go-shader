package main

import "math"

type Vector2 struct {
	X float64
	Y float64
}

var Vector3Left Vector2 = Vector2{
	X: -1.0,
	Y: 0.0,
}

var Vector3Right Vector2 = Vector2{
	X: 1.0,
	Y: 0.0,
}

var Vector3Up Vector2 = Vector2{
	X: 0.0,
	Y: 1.0,
}

var Vector3Down Vector2 = Vector2{
	X: 0.0,
	Y: -1.0,
}

func (self Vector2) Add(other Vector2) Vector2 {
	return Vector2{
		X: self.X + other.X,
		Y: self.Y + other.Y,
	}
}

func (self Vector2) Sub(other Vector2) Vector2 {
	return Vector2{
		X: self.X - other.X,
		Y: self.Y - other.Y,
	}
}

func (self Vector2) Mul(scalar float64) Vector2 {
	return Vector2{
		X: self.X * scalar,
		Y: self.Y * scalar,
	}
}

func (self Vector2) Div(scalar float64) Vector2 {
	if scalar == 0.0 {
		return self
	}

	return Vector2{
		X: self.X / scalar,
		Y: self.Y / scalar,
	}
}

func (self Vector2) Magnitude() float64 {
	return math.Sqrt(self.X*self.X + self.Y*self.Y)
}

func (self Vector2) Normalize() Vector2 {
	return self.Div(self.Magnitude())
}

func (self Vector2) Reflect(normal Vector2) Vector2 {
	return self.Sub(normal.Mul(2 * self.Dot(normal)))
}

func (self Vector2) Turn(degrees float64) Vector2 {
	theta := Deg2Rad(degrees)
	cs := math.Cos(theta)
	sn := math.Sin(theta)
	vx := self.X*cs - self.Y*sn
	vy := self.X*sn + self.Y*cs
	return Vector2{
		X: vx,
		Y: vy,
	}
}

func (self Vector2) Dot(other Vector2) float64 {
	return self.X*other.X + self.Y*other.Y
}

func Deg2Rad(degrees float64) float64 {
	return degrees * math.Pi / 180
}
