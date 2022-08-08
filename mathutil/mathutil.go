package mathutil

import (
	"math"
	"math/rand"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
)

var PI float64 = 3.14159265358979323846

var Vector3Left mgl64.Vec2 = mgl64.Vec2{-1.0, 0.0}

var Vector3Right mgl64.Vec2 = mgl64.Vec2{1.0, 0.0}

var Vector3Up mgl64.Vec2 = mgl64.Vec2{0.0, 1.0}

var Vector3Down mgl64.Vec2 = mgl64.Vec2{0.0, -1.0}

func ReflectVec2(self, normal mgl64.Vec2) mgl64.Vec2 {
	return self.Sub(normal.Mul(2 * self.Dot(normal)))
}

func TurnVec2(self mgl64.Vec2, degrees float64) mgl64.Vec2 {
	theta := Deg2Rad(degrees)
	cs := math.Cos(theta)
	sn := math.Sin(theta)
	vx := self[0]*cs - self[1]*sn
	vy := self[0]*sn + self[1]*cs
	return mgl64.Vec2{vx, vy}
}

func Deg2Rad32(degrees float32) float32 {
	return degrees * math.Pi / 180
}

func Deg2Rad(degrees float64) float64 {
	return degrees * math.Pi / 180
}

func RandSign() float32 {
	x := rand.Float32()
	if x < 0.5 {
		return 0.0
	}

	return 1.0
}

func RandMGL32Vec3() mgl32.Vec3 {
	return mgl32.Vec3{
		float32(rand.Float64()),
		float32(rand.Float64()),
		float32(rand.Float64()),
	}
}

// rand sphere point generation https://karthikkaranth.me/blog/generating-random-points-in-a-sphere/
func RandPointInSphere[T float32 | float64](radius float64) (T, T, T) {
	u := rand.Float64()
	v := rand.Float64()
	theta := u * 2.0 * math.Pi
	phi := math.Acos(2.0*v - 1.0)
	r := math.Cbrt(rand.Float64())
	sinTheta := math.Sin(theta)
	cosTheta := math.Cos(theta)
	sinPhi := math.Sin(phi)
	cosPhi := math.Cos(phi)

	x := T(r * sinPhi * cosTheta * radius)
	y := T(r * sinPhi * sinTheta * radius)
	z := T(r * cosPhi * radius)

	return x, y, z
}
